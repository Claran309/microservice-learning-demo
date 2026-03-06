package main

import (
	"log"
	"microservicesDemo/L3-etcd/internal/user-service/dao"
	"microservicesDemo/L3-etcd/internal/user-service/handler"
	"microservicesDemo/L3-etcd/internal/user-service/service"
	"microservicesDemo/L3-etcd/kitex_gen/user/userservice"
	"net"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {
	db, err := dao.InitDB("claran:chr070309@tcp(localhost:3306)/microserviceDemo?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
	}

	userRepo := dao.NewUserRepo(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserServiceImpl(userService)

	// 启动etcd注册中心
	r, err := etcd.NewEtcdRegistry([]string{"localhost:2379"})
	if err != nil {
		log.Fatal(err)
	}

	// 创建地址
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	svr := userservice.NewServer(userHandler,
		server.WithServiceAddr(addr),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "user-service",
		}),
		server.WithRegistry(r))

	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
