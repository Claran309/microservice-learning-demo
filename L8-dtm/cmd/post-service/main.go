package main

import (
	"context"
	"microservicesDemo/L8-dtm/internal/post-service/dao"
	"microservicesDemo/L8-dtm/internal/post-service/handler"
	"microservicesDemo/L8-dtm/internal/post-service/service"
	"microservicesDemo/L8-dtm/kitex_gen/post/postservice"
	"microservicesDemo/L8-dtm/pkg/cache/redis"
	"microservicesDemo/L8-dtm/pkg/id/snowflake"
	log "microservicesDemo/L8-dtm/pkg/logging/zap"
	mq "microservicesDemo/L8-dtm/pkg/mq/kafka"
	"net"

	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	etcd "github.com/kitex-contrib/registry-etcd"
	"go.uber.org/zap"
)

func main() {
	log.InitLogManager("post-service", "./logs/post-service")
	zap.L().Info("√ 初始化Zap成功")

	if err := snowflake.InitSnowflake(2); err != nil {
		zap.L().Fatal("× 初始化雪花算法失败", zap.Error(err))
	}
	zap.L().Info("√ 初始化雪花算法ID生成器成功")

	db, err := dao.InitDB("claran:chr070309@tcp(:3306)/microserviceDemo?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		zap.L().Fatal("× 初始化数据库失败", zap.Error(err))
	}
	zap.L().Info("√ 初始化数据库成功")

	redisCluster, err := redis.NewRedisCluster(redis.RedisConfig{
		Addrs:        []string{":7000", ":7001", ":7002", ":7003", ":7004", ":7005"},
		Password:     "",
		PoolSize:     100,
		MinIdleConns: 10,
		DialTimeout:  5 * 1000000000,
		ReadTimeout:  3 * 1000000000,
		WriteTimeout: 3 * 1000000000,
	})
	if err != nil {
		zap.L().Fatal("× 初始化Redis集群失败", zap.Error(err))
	}
	zap.L().Info("√ 初始化Redis集群成功")

	kafkaProducer, err := mq.NewProducer()
	if err != nil {
		zap.L().Fatal("× 初始化Kafka生产者失败", zap.Error(err))
	}
	zap.L().Info("√ 初始化Kafka生产者成功")

	otelProvider := provider.NewOpenTelemetryProvider(
		provider.WithServiceName("post-service"),
		provider.WithExportEndpoint(":8079"),
		provider.WithInsecure(),
		provider.WithEnableTracing(true),
		provider.WithEnableMetrics(true),
	)
	defer otelProvider.Shutdown(context.Background())
	zap.L().Info("√ OpenTelemetry 初始化完成")

	postRepo := dao.NewPostRepositoryImpl(db, redisCluster)
	zap.L().Info("√ 初始化仓储层成功")

	postService := service.NewPostService(postRepo, kafkaProducer)
	zap.L().Info("√ 初始化服务层成功")

	postHandler := handler.NewPostServiceImpl(postService)
	zap.L().Info("√ 初始化处理器成功")

	r, err := etcd.NewEtcdRegistry([]string{":2379"})
	if err != nil {
		zap.L().Fatal("× 初始化etcd注册中心失败", zap.Error(err))
	}
	zap.L().Info("√ 初始化etcd注册中心成功")

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8081")
	if err != nil {
		zap.L().Fatal("× 解析TCP地址失败", zap.Error(err))
	}
	zap.L().Info("√ 解析TCP地址成功")

	svr := postservice.NewServer(postHandler,
		server.WithServiceAddr(addr),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "post-service",
		}),
		server.WithRegistry(r),
		server.WithLimit(&limit.Option{
			MaxConnections: 1000,
			MaxQPS:         1000,
		}),
		server.WithSuite(tracing.NewServerSuite()),
	)
	zap.L().Info("√ 创建Kitex服务器成功")

	zap.L().Info("√ post-service 启动成功，开始监听 127.0.0.1:8081")
	if err := svr.Run(); err != nil {
		zap.L().Fatal("× 服务器运行失败", zap.Error(err))
	}
}
