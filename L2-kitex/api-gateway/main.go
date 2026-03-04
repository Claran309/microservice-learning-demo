package main

import (
	"log"
	"microservicesDemo/L2-kitex/kitex_gen/user/userservice"

	"github.com/cloudwego/hertz/pkg/app/server" // Hertz服务器
	"github.com/cloudwego/kitex/client"         // Kitex客户端
)

// 全局RPC客户端
var userClient userservice.Client

// init函数，在main之前执行
func init() {
	log.Println("初始化rpc客户端...")

	// 创建kitex客户端
	var err error
	userClient, err = userservice.NewClient("user-service", client.WithHostPorts("127.0.0.1:8080"))

	if err != nil {
		log.Fatal("创建客户端失败", err)
	}

	log.Println("RPC客户端初始化成功")
	log.Println("连接到: 127.0.0.1:8888")
}

func main() {
	h := server.Default(server.WithHostPorts(":8888"))

	handler := APIHandler{}

	RegisterRoutes(h, handler)

	log.Printf("网关服务器正在端口 8888 上启动")
	h.Spin()
}
