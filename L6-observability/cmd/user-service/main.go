package main

import (
	"context"
	"log"
	"microservicesDemo/L6-observability/internal/user-service/dao"
	"microservicesDemo/L6-observability/internal/user-service/handler"
	"microservicesDemo/L6-observability/internal/user-service/service"
	"microservicesDemo/L6-observability/kitex_gen/user/userservice"
	mq "microservicesDemo/L6-observability/pkg/mq/kafka"
	"net"

	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {
	// 初始化数据库
	db, err := dao.InitDB("claran:chr070309@tcp(localhost:3306)/microserviceDemo?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
	}

	// 初始化Kafka
	kafkaProducer, err := mq.NewProducer()
	if err != nil {
		log.Fatal(err)
	}

	// 初始化 OpenTelemetry provider
	otelProvider := provider.NewOpenTelemetryProvider(
		provider.WithServiceName("user-service"),
		provider.WithExportEndpoint("localhost:8079"),
		provider.WithInsecure(),
		provider.WithEnableTracing(true),
		provider.WithEnableMetrics(true),
	)
	defer otelProvider.Shutdown(context.Background())

	log.Println("OpenTelemetry 初始化完成")

	userRepo := dao.NewUserRepo(db)
	userService := service.NewUserService(userRepo, kafkaProducer)
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
		server.WithRegistry(r),
		server.WithLimit(&limit.Option{
			MaxConnections: 1000,
			MaxQPS:         1000,
		}),
		server.WithSuite(tracing.NewServerSuite()), // 使用 Kitex 的 tracing
	)

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
