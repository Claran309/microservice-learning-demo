package main

import (
	"context"
	"log"
	"time"
	"user-service/kitex_gen/user"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

type APIHandler struct {
}

func NewAPIHandler() *APIHandler {
	return &APIHandler{}
}

func (h *APIHandler) Register(c context.Context, ctx *app.RequestContext) {
	// 常规handler
	startTime := time.Now()

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := ctx.BindAndValidate(&req); err != nil {
		log.Println("绑定参数时发生错误", err)
		ctx.JSON(400, utils.H{
			"code": 400,
			"msg":  "绑定参数时发生错误",
		})
		return
	}

	// 构造rpc请求，包名为构建时填写的包名
	rpcReq := &user.RegisterReq{
		Username: req.Username,
		Password: req.Password,
	}

	// 调用rpc请求
	log.Println("调用rpc服务...")
	rpcResp, err := userClient.Register(c, rpcReq)
	if err != nil {
		log.Printf("RPC调用失败: %v", err)
		ctx.JSON(500, utils.H{
			"code":    500,
			"message": "服务内部错误: " + err.Error(),
		})
		return
	}

	// 判断rpc服务成功与否
	statusCode := 200
	if !rpcResp.Success {
		statusCode = 400
	}

	duration := time.Since(startTime)

	// 返回HTTP响应
	log.Printf("注册处理完成，状态码: %d，耗时: %v", statusCode, duration)

	ctx.JSON(statusCode, utils.H{
		"code":    statusCode,      // HTTP状态码
		"success": rpcResp.Success, // 业务是否成功
		"message": rpcResp.Message, // 业务消息
	})
}
