package main

import (
	"microservicesDemo/L4-govern/internal/api-gateway/client"
	api_gateway2 "microservicesDemo/L4-govern/internal/api-gateway/handler"
	"microservicesDemo/L4-govern/internal/api-gateway/router"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	// 创建rpc客户端
	clients := client.InitClients()

	// 创建http处理器
	httpHandler := api_gateway2.NewHttpHandler(clients)

	// 启动hertz服务器
	h := server.Default()

	// 注册路由
	router.InitRouters(h, httpHandler)

	h.Spin()
}
