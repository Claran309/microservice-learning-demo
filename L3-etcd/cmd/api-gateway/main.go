package main

import (
	api_gateway2 "microservicesDemo/L3-etcd/internal/api-gateway"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	// 创建rpc客户端
	clients := api_gateway2.InitClients()

	// 创建http处理器
	httpHandler := api_gateway2.NewHttpHandler(clients)

	// 启动hertz服务器
	h := server.Default()

	// 注册路由
	api_gateway2.InitRouters(h, httpHandler)

	h.Spin()
}
