package main

import (
	"context"
	"microservicesDemo/L8-dtm/internal/api-gateway/client"
	api_gateway2 "microservicesDemo/L8-dtm/internal/api-gateway/handler"
	"microservicesDemo/L8-dtm/internal/api-gateway/router"
	"microservicesDemo/L8-dtm/pkg/cache/redis"
	"microservicesDemo/L8-dtm/pkg/dtm"
	"microservicesDemo/L8-dtm/pkg/id/snowflake"
	"microservicesDemo/L8-dtm/pkg/logging/zap"

	"github.com/cloudwego/hertz/pkg/app/server"
	hertztracing "github.com/hertz-contrib/obs-opentelemetry/provider"
	"github.com/hertz-contrib/obs-opentelemetry/tracing"
	"go.uber.org/zap"
)

func main() {
	log.InitLogManager("api-gateway", "./logs/api-gateway")

	if err := snowflake.InitSnowflake(0); err != nil {
		zap.L().Fatal("× 初始化雪花算法失败", zap.Error(err))
	}
	zap.L().Info("√ 初始化雪花算法ID生成器成功")

	p := hertztracing.NewOpenTelemetryProvider(
		hertztracing.WithServiceName("api-gateway"),
		hertztracing.WithExportEndpoint("localhost:8079"),
		hertztracing.WithInsecure(),
	)
	defer p.Shutdown(context.Background())
	zap.L().Info("√ 初始化Hertz OTel Provider成功")

	redisCluster, err := redis.NewRedisCluster(redis.RedisConfig{
		Addrs:        []string{":7000", ":7001", ":7002", ":7003", ":7004", ":7005"},
		Password:     "",
		PoolSize:     100,
		MinIdleConns: 10,
	})
	if err != nil {
		zap.L().Fatal("× 初始化Redis集群失败", zap.Error(err))
	}
	zap.L().Info("√ 初始化Redis集群成功")

	dtmManager := dtm.NewDTMManager("http://localhost:36789")
	zap.L().Info("√ 初始化DTM管理器成功")

	tracer, cfg := tracing.NewServerTracer()

	clients := client.InitClients()
	zap.L().Info("√ 创建rpc客户端成功")

	httpHandler := api_gateway2.NewHttpHandler(clients, redisCluster, dtmManager)
	zap.L().Info("√ 创建http处理器成功")

	h := server.Default(tracer)
	zap.L().Info("√ 启动hertz服务器成功")

	h.Use(tracing.ServerMiddleware(cfg))

	router.InitRouters(h, httpHandler)
	zap.L().Info("√ 注册路由成功")

	h.Spin()
}
