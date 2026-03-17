package main

import (
	"context"
	"log"
	"microservicesDemo/L6-observability/internal/post-service/dao"
	"microservicesDemo/L6-observability/internal/post-service/handler"
	"microservicesDemo/L6-observability/internal/post-service/service"
	"microservicesDemo/L6-observability/kitex_gen/post/postservice"
	mq "microservicesDemo/L6-observability/pkg/mq/kafka"
	"net"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
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

	// 初始化 OpenTelemetry provider
	otelProvider := provider.NewOpenTelemetryProvider(
		provider.WithServiceName("post-service"),
		provider.WithExportEndpoint("localhost:8079"),
		provider.WithInsecure(),
		provider.WithEnableTracing(true),
		provider.WithEnableMetrics(true),
	)
	defer otelProvider.Shutdown(context.Background())

	log.Println("OpenTelemetry 初始化完成")

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
		server.WithSuite(tracing.NewServerSuite()), // 使用 Kitex 的 tracing
		server.WithRegistry(r))

	log.Println("启动 user-service 在 127.0.0.1:8080")
	log.Println("OpenTelemetry Collector: localhost:8079")
	log.Println("监控面板:")
	log.Println("  Jaeger: http://localhost:8085")
	log.Println("  Prometheus: http://localhost:8084")
	log.Println("  Grafana: http://localhost:8086 (admin/admin123)")

	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
