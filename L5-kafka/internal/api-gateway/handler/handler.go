package handler

import (
	"context"
	"log"
	client2 "microservicesDemo/L5-kafka/internal/api-gateway/client"
	"microservicesDemo/L5-kafka/kitex_gen/user"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

type HttpHandler struct {
	client client2.Clients
}

func NewHttpHandler(client client2.Clients) *HttpHandler {
	return &HttpHandler{
		client: client,
	}
}

func (h *HttpHandler) Register(c context.Context, ctx *app.RequestContext) {
	startTime := time.Now()

	var req struct {
		Name     string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	if err := ctx.BindAndValidate(&req); err != nil {
		log.Println("绑定参数时发生错误", err)
		ctx.JSON(400, utils.H{
			"code": 400,
			"msg":  "绑定参数时发生错误",
		})
		return
	}

	rpcReq := user.RegisterReq{
		Username: req.Name,
		Password: req.Password,
		Email:    req.Email,
	}

	rpcResp, err := h.client.UserClient.Register(c, &rpcReq)
	if err != nil {
		log.Printf("RPC调用失败: %v", err)
		ctx.JSON(500, utils.H{
			"code":    500,
			"message": "服务内部错误: " + err.Error(),
		})
		return
	}

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
		"message": rpcResp.Msg,     // 业务消息
	})
}

func (h *HttpHandler) Login(c context.Context, ctx *app.RequestContext) {
	startTime := time.Now()

	var req struct {
		Name     string `json:"username"`
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

	rpcReq := user.LoginByUsernameReq{
		Username: req.Name,
		Password: req.Password,
	}

	rpcResp, err := h.client.UserClient.Login(c, &rpcReq)
	if err != nil {
		log.Printf("RPC调用失败: %v", err)
		ctx.JSON(500, utils.H{
			"code":    500,
			"message": "服务内部错误: " + err.Error(),
		})
		return
	}

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
		"message": rpcResp.Msg,     // 业务消息
	})
}

func (h *HttpHandler) CreatePost(c context.Context, ctx *app.RequestContext) {}

func (h *HttpHandler) DeletePost(c context.Context, ctx *app.RequestContext) {}
