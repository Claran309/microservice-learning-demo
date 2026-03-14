package main

import (
	"log"
	"microservicesDemo/L5-kafka/internal/post-service/dao"
	"microservicesDemo/L5-kafka/internal/post-service/handler"
	"microservicesDemo/L5-kafka/internal/post-service/service"
	"microservicesDemo/L5-kafka/kitex_gen/post/postservice"
	mq "microservicesDemo/L5-kafka/pkg/mq/kafka"
	"net"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {
	db, err := dao.InitDB("claran:chr070309@tcp(localhost:3306)/microserviceDemo?charset=utf8mb4&parseTime=True&loc=Local")

	// 初始化消息队列
	kafkaProducer, err := mq.NewProducer()
	if err != nil {
		log.Fatal(err)
	}
	defer kafkaProducer.Close()

	kafkaConsumer, err := mq.NewConsumer("user", "1")
	if err != nil {
		log.Fatal(err)
	}
	defer kafkaConsumer.Close()

	postRepo := dao.NewPostRepositoryImpl(db)
	postService := service.NewPostServiceImpl(postRepo, kafkaProducer, kafkaConsumer)
	postHandler := handler.NewPostServiceImpl(postService)

	// 启动消费者监听
	postService.StartConsumer("user", "1")
	defer postService.StopConsumer()

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
