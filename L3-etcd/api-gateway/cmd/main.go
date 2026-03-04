package cmd

import (
	"microservicesDemo/L3-etcd/api-gateway/internal"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	// 创建rpc客户端
	clients := internal.InitClients()

	// 创建http处理器
	httpHandler := internal.NewHttpHandler(clients)

	// 启动hertz服务器
	h := server.Default()

	// 注册路由
	internal.InitRouters(h, httpHandler)
}
