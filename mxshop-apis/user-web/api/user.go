package api

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"mxshop_api/user-web/forms"
	"mxshop_api/user-web/global"
	"mxshop_api/user-web/global/response"
	"mxshop_api/user-web/middlewares"
	"mxshop_api/user-web/models"
	"mxshop_api/user-web/proto"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func removeTopStruct(fileds map[string]string) map[string]string {
	rsp := map[string]string{}
	for field, err := range fileds {
		rsp[field[strings.Index(field, ".")+1:]] = err
	}
	return rsp
}
func HandleValidatorError(c *gin.Context, err error) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"msg": err.Error(),
		})
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"error": removeTopStruct(errs.Translate(global.Trans)),
	})
	return
}
func HandleGrpcErrorToHttp(err error, c *gin.Context) {
	//将grpc的code转换为http的状态码
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"msg": e.Message(),
				})
			case codes.Internal:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "内部错误",
				})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{
					"msg": "参数错误",
				})
			case codes.Unavailable:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "用户服务不可用",
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "其他错误",
				})
			}
			return
		}
	}
}
func GetUserList(c *gin.Context) {
	claims, _ := c.Get("claims")
	currentUser := claims.(*models.CustomClaims)
	zap.S().Infof("访问用户:%d", currentUser.ID)
	pn := c.DefaultQuery("pn", "0")
	pnInt, _ := strconv.Atoi(pn)
	pSize := c.DefaultQuery("psize", "5")
	pSizeInt, _ := strconv.Atoi(pSize)
	rsp, err := global.UserSrvClient.GetUserList(context.Background(), &proto.PageInfo{
		Pn:    uint32(pnInt),
		PSize: uint32(pSizeInt),
	})
	if err != nil {
		zap.S().Errorw("[GetUserList] 查询 [用户列表] 失败")
		HandleGrpcErrorToHttp(err, c)
		return
	}
	result := make([]interface{}, 0)
	for _, value := range rsp.Data {
		//data :=make(map[string]interface{})
		user := response.UserResponse{
			Id:       value.Id,
			NickName: value.NickName,
			BirthDay: time.Time(time.Unix(int64(value.BirthDay), 0)).Format("2006-01-02"),
			Gender:   value.Gender,
			Mobile:   value.Mobile,
		}
		result = append(result, user)
	}
	c.JSON(http.StatusOK, result)
}
func PassWordLogin(c *gin.Context) {
	passwordLoginForm := forms.PassWordLoginForm{}
	if err := c.ShouldBind(&passwordLoginForm); err != nil {
		HandleValidatorError(c, err)
		return
	}
	if !store.Verify(passwordLoginForm.CaptchaId, passwordLoginForm.Captcha, true) {
		c.JSON(http.StatusBadRequest, gin.H{
			"captcha": "验证码错误",
		})
		return
	}
	//登录的逻辑
	if rsp, err := global.UserSrvClient.GetUserByMobile(context.Background(), &proto.MobileRequest{
		Mobile: passwordLoginForm.Mobile,
	}); err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				c.JSON(http.StatusBadRequest, map[string]string{
					"mobile": "用户不存在",
				})
			default:
				c.JSON(http.StatusInternalServerError, map[string]string{
					"mobile": "登录失败2",
				})
			}
			return
		}
	} else {
		//只查询的用户，没有检查密码
		if passRsp, pasErr := global.UserSrvClient.CheckPassWord(context.Background(), &proto.PasswordCheckInfo{
			Password:          passwordLoginForm.PassWord,
			EncryptedPassword: rsp.PassWord,
		}); pasErr != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{
				"password": "登录失败",
			})
		} else {
			if passRsp.Success {
				//生成token
				j := middlewares.NewJWT()
				claims := models.CustomClaims{
					ID:          uint(rsp.Id),
					NickName:    rsp.NickName,
					AuthorityId: uint(rsp.Role),
					StandardClaims: jwt.StandardClaims{
						NotBefore: time.Now().Unix(),               //当前时间生效
						ExpiresAt: time.Now().Unix() + 60*60*24*30, //30天过期
						Issuer:    "wam",
					},
				}
				token, err := j.CreateToken(claims)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"msg": "生成token失败",
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"id":         rsp.Id,
					"nickname":   rsp.NickName,
					"token":      token,
					"expired_at": (time.Now().Unix() + 60*60*24*30) * 1000,
				})
			} else {
				c.JSON(http.StatusBadRequest, map[string]string{
					"msg": "登录失败",
				})
			}
		}
	}
}
func Register(c *gin.Context) {
	registerForm := forms.RegisterForm{}
	if err := c.ShouldBind(&registerForm); err != nil {
		HandleValidatorError(c, err)
		return
	}
	user, err := global.UserSrvClient.CreateUser(context.Background(), &proto.CreateUserInfo{
		NickName: registerForm.Mobile,
		PassWord: registerForm.PassWord,
		Mobile:   registerForm.Mobile,
	})
	if err != nil {
		zap.S().Errorf("[Register] [新建用户] 失败:%s", err.Error())
		HandleGrpcErrorToHttp(err, c)
		return
	}
	j := middlewares.NewJWT()
	claims := models.CustomClaims{
		ID:          uint(user.Id),
		NickName:    user.NickName,
		AuthorityId: uint(user.Role),
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),               //当前时间生效
			ExpiresAt: time.Now().Unix() + 60*60*24*30, //30天过期
			Issuer:    "wam",
		},
	}
	token, err := j.CreateToken(claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "生成token失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":         user.Id,
		"nickname":   user.NickName,
		"token":      token,
		"expired_at": (time.Now().Unix() + 60*60*24*30) * 1000,
	})
}
