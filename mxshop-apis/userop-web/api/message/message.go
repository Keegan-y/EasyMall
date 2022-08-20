package message

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mxshop_api/userop-web/api"
	"mxshop_api/userop-web/forms"
	"mxshop_api/userop-web/global"
	"mxshop_api/userop-web/models"
	"mxshop_api/userop-web/proto"
	"net/http"
)

func List(c *gin.Context) {
	request := &proto.MessageRequest{}
	userId, _ := c.Get("userId")
	claims, _ := c.Get("claims")
	model := claims.(*models.CustomClaims)
	if model.AuthorityId == 1 {
		request.UserId = int32(userId.(uint))
	}

	rsp, err := global.MessageClient.MessageList(context.Background(), request)
	if err != nil {
		zap.S().Errorw("获取留言失败")
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	reMap := map[string]interface{}{
		"total": rsp.Total,
	}
	result := make([]interface{}, 0)
	for _, value := range rsp.Data {
		reMap := make(map[string]interface{})
		reMap["id"] = value.Id
		reMap["user_id"] = value.UserId
		reMap["type"] = value.MessageType
		reMap["subject"] = value.Subject
		reMap["message"] = value.Message
		reMap["file"] = value.File

		result = append(result, reMap)
	}
	reMap["data"] = result

	c.JSON(http.StatusOK, reMap)
}

func New(c *gin.Context) {
	userId, _ := c.Get("userId")

	messageForm := forms.MessageForm{}
	if err := c.ShouldBindJSON(&messageForm); err != nil {
		api.HandleValidatorError(c, err)
		return
	}

	rsp, err := global.MessageClient.CreateMessage(context.Background(), &proto.MessageRequest{
		UserId: int32(userId.(uint)),
		MessageType: messageForm.MessageType,
		Subject: messageForm.Subject,
		Message: messageForm.Message,
		File: messageForm.File,
	})

	if err != nil {
		zap.S().Errorw("添加留言失败")
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":rsp.Id,
	})
}