package main

import (
	"log"
	"microservicesDemo/L3-etcd/internal/post-service/dao"
	"microservicesDemo/L3-etcd/internal/post-service/handler"
	"microservicesDemo/L3-etcd/internal/post-service/service"
	"microservicesDemo/L3-etcd/kitex_gen/post/postservice"
	"net"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {
	db, err := dao.InitDB("claran:chr070309@tcp(localhost:3306)/microserviceDemo?charset=utf8mb4&parseTime=True&loc=Local")

	postRepo := dao.NewPostRepositoryImpl(db)
	postService := service.NewPostServiceImpl(postRepo)
	postHandler := handler.NewPostServiceImpl(postService)

	r, err := etcd.NewEtcdRegistry([]string{"localhost:2379"})

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8081")
	if err != nil {
		panic(err)
	}

	svr := postservice.NewServer(postHandler,
		server.WithServiceAddr(addr),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "post-service",
		}),
		server.WithRegistry(r))

	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
