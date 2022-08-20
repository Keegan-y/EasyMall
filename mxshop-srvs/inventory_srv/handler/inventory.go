package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"mxshop_srvs/inventory_srv/global"
	"mxshop_srvs/inventory_srv/model"
	"mxshop_srvs/inventory_srv/proto"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/types/known/emptypb"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer
}

func (*InventoryServer) SetInv(c context.Context, req *proto.GoodsInvInfo) (*emptypb.Empty, error) {
	//设置库存
	var inv model.Inventory
	global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv)
	inv.Goods = req.GoodsId
	inv.Stocks = req.Num
	global.DB.Save(&inv)
	return &emptypb.Empty{}, nil
}

func (*InventoryServer) InvDetail(c context.Context, req *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	var inv model.Inventory
	if result := global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "没有库存信息")
	}
	return &proto.GoodsInvInfo{
		GoodsId: inv.Goods,
		Num:     inv.Stocks,
	}, nil
}

func (*InventoryServer) Sell(c context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//扣减库存
	//并发情况下会出现超卖,要靠redis分布式锁来解决
	client := global.RDB
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
	rs := redsync.New(pool)
	//开启事务
	tx := global.DB.Begin()

	sellDetail := model.StockSellDetail{
		OrderSn: req.OrderSn,
		Status:  1,
	}
	var details []model.GoodsDetail
	for _, goodInfo := range req.GoodsInfo {
		details = append(details, model.GoodsDetail{
			Goods: goodInfo.GoodsId,
			Num:   goodInfo.Num,
		})
		var inv model.Inventory
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常")
		}
		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		//判断库存是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		//扣减库存
		inv.Stocks -= goodInfo.Num
		tx.Save(&inv)
		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
	}
	sellDetail.Detail = details
	//写selldetail表
	if result := tx.Create(&sellDetail); result.RowsAffected == 0 {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "保存库存扣减历史失败")
	}
	tx.Commit() //提交事务
	return &emptypb.Empty{}, nil
}

func (*InventoryServer) Reback(c context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	/*库存归还:
	1.订单超时归还
	2.订单创建失败，归还之前扣减的商品
	3.手动归还
	*/
	tx := global.DB.Begin()
	for _, goodInfo := range req.GoodsInfo {
		var inv model.Inventory
		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}

		//扣减， 会出现数据不一致的问题
		inv.Stocks += goodInfo.Num
		tx.Save(&inv)
	}
	tx.Commit() // 需要自己手动提交操作
	return &emptypb.Empty{}, nil
}

func (*InventoryServer) TrySell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//扣减库存， 本地事务 [1:10,  2:5, 3: 20]
	//数据库基本的一个应用场景：数据库事务
	//并发情况之下 可能会出现超卖 1
	client := global.RDB
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
	rs := redsync.New(pool)

	tx := global.DB.Begin()
	//m.Lock() //获取锁 这把锁有问题吗？  假设有10w的并发， 这里并不是请求的同一件商品  这个锁就没有问题了吗？
	for _, goodInfo := range req.GoodsInfo {
		var inv model.InventoryNew
		//if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods:goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
		//	tx.Rollback() //回滚之前的操作
		//	return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		//}

		//for {
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常")
		}

		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		//判断库存是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		//扣减， 会出现数据不一致的问题 - 锁，分布式锁
		//inv.Stocks -= goodInfo.Num
		inv.Freeze += goodInfo.Num
		tx.Save(&inv)

		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
		//update inventory set stocks = stocks-1, version=version+1 where goods=goods and version=version
		//这种写法有瑕疵，为什么？
		//零值 对于int类型来说 默认值是0 这种会被gorm给忽略掉
		//if result := tx.Model(&model.Inventory{}).Select("Stocks", "Version").Where("goods = ? and version= ?", goodInfo.GoodsId, inv.Version).Updates(model.Inventory{Stocks: inv.Stocks, Version: inv.Version+1}); result.RowsAffected == 0 {
		//	zap.S().Info("库存扣减失败")
		//}else{
		//	break
		//}
		//}
		//tx.Save(&inv)
	}
	tx.Commit() // 需要自己手动提交操作
	//m.Unlock() //释放锁
	return &emptypb.Empty{}, nil
}

func (*InventoryServer) ConfirmSell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//扣减库存， 本地事务 [1:10,  2:5, 3: 20]
	//数据库基本的一个应用场景：数据库事务
	//并发情况之下 可能会出现超卖 1
	client := global.RDB
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
	rs := redsync.New(pool)

	tx := global.DB.Begin()
	//m.Lock() //获取锁 这把锁有问题吗？  假设有10w的并发， 这里并不是请求的同一件商品  这个锁就没有问题了吗？
	for _, goodInfo := range req.GoodsInfo {
		var inv model.InventoryNew
		//if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods:goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
		//	tx.Rollback() //回滚之前的操作
		//	return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		//}

		//for {
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常")
		}

		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		//判断库存是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		//扣减， 会出现数据不一致的问题 - 锁，分布式锁
		inv.Stocks -= goodInfo.Num
		inv.Freeze -= goodInfo.Num
		tx.Save(&inv)

		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
		//update inventory set stocks = stocks-1, version=version+1 where goods=goods and version=version
		//这种写法有瑕疵，为什么？
		//零值 对于int类型来说 默认值是0 这种会被gorm给忽略掉
		//if result := tx.Model(&model.Inventory{}).Select("Stocks", "Version").Where("goods = ? and version= ?", goodInfo.GoodsId, inv.Version).Updates(model.Inventory{Stocks: inv.Stocks, Version: inv.Version+1}); result.RowsAffected == 0 {
		//	zap.S().Info("库存扣减失败")
		//}else{
		//	break
		//}
		//}
		//tx.Save(&inv)
	}
	tx.Commit() // 需要自己手动提交操作
	//m.Unlock() //释放锁
	return &emptypb.Empty{}, nil
}

func (*InventoryServer) CancelSell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//扣减库存， 本地事务 [1:10,  2:5, 3: 20]
	//数据库基本的一个应用场景：数据库事务
	//并发情况之下 可能会出现超卖 1
	client := global.RDB
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
	rs := redsync.New(pool)

	tx := global.DB.Begin()
	//m.Lock() //获取锁 这把锁有问题吗？  假设有10w的并发， 这里并不是请求的同一件商品  这个锁就没有问题了吗？
	for _, goodInfo := range req.GoodsInfo {
		var inv model.InventoryNew
		//if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods:goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
		//	tx.Rollback() //回滚之前的操作
		//	return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		//}

		//for {
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常")
		}

		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		//判断库存是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		//扣减， 会出现数据不一致的问题 - 锁，分布式锁
		inv.Freeze -= goodInfo.Num
		tx.Save(&inv)

		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
		//update inventory set stocks = stocks-1, version=version+1 where goods=goods and version=version
		//这种写法有瑕疵，为什么？
		//零值 对于int类型来说 默认值是0 这种会被gorm给忽略掉
		//if result := tx.Model(&model.Inventory{}).Select("Stocks", "Version").Where("goods = ? and version= ?", goodInfo.GoodsId, inv.Version).Updates(model.Inventory{Stocks: inv.Stocks, Version: inv.Version+1}); result.RowsAffected == 0 {
		//	zap.S().Info("库存扣减失败")
		//}else{
		//	break
		//}
		//}
		//tx.Save(&inv)
	}
	tx.Commit() // 需要自己手动提交操作
	//m.Unlock() //释放锁
	return &emptypb.Empty{}, nil
}

func AutoReback(c context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	type OrderInfo struct {
		OrderSn string
	}
	for i := range msgs {
		//为避免重复归还的问题,这个个接口应该确保幂等性
		var orderInfo OrderInfo
		err := json.Unmarshal(msgs[i].Body, &orderInfo)
		if err != nil {
			zap.S().Errorf("解析json失败： %v\n", msgs[i].Body)
			return consumer.ConsumeSuccess, nil
		}

		//去将inv的库存加回去 将selldetail的status设置为2， 要在事务中进行
		tx := global.DB.Begin()
		var sellDetail model.StockSellDetail
		if result := tx.Model(&model.StockSellDetail{}).Where(&model.StockSellDetail{OrderSn: orderInfo.OrderSn, Status: 1}).First(&sellDetail); result.RowsAffected == 0 {
			return consumer.ConsumeSuccess, nil
		}
		//如果查询到那么逐个归还库存
		for _, orderGood := range sellDetail.Detail {
			if result := tx.Model(&model.Inventory{}).Where(&model.Inventory{Goods: orderGood.Goods}).Update("stocks", gorm.Expr("stocks+?", orderGood.Num)); result.RowsAffected == 0 {
				tx.Rollback()
				return consumer.ConsumeRetryLater, nil
			}
		}

		if result := tx.Model(&model.StockSellDetail{}).Where(&model.StockSellDetail{OrderSn: orderInfo.OrderSn}).Update("status", 2); result.RowsAffected == 0 {
			tx.Rollback()
			return consumer.ConsumeRetryLater, nil
		}
		tx.Commit()
		return consumer.ConsumeSuccess, nil
	}
	return consumer.ConsumeSuccess, nil
}
