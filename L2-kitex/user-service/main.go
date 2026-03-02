package main

import (
	"net"
	"user-service/kitex_gen/user/userservice"

	"github.com/cloudwego/kitex/server"
)

func main() {
	// 创建地址
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	svr := userservice.NewServer(
		new(UserServiceImpl),         // 业务处理器
		server.WithServiceAddr(addr), // 监听端口
	)

	//启动rpc服务
	err = svr.Run()
	if err != nil {
		panic(err)
	}
}
