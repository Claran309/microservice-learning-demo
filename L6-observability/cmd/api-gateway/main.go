package main

import (
	"context"
	"microservicesDemo/L6-observability/internal/api-gateway/client"
	api_gateway2 "microservicesDemo/L6-observability/internal/api-gateway/handler"
	"microservicesDemo/L6-observability/internal/api-gateway/router"

	"github.com/cloudwego/hertz/pkg/app/server"
	hertztracing "github.com/hertz-contrib/obs-opentelemetry/provider"
	"github.com/hertz-contrib/obs-opentelemetry/tracing"
)

func main() {
	// 初始化OTel Provider
	p := hertztracing.NewOpenTelemetryProvider(
		hertztracing.WithServiceName("api-gateway"),
		hertztracing.WithExportEndpoint("localhost:8079"),
		hertztracing.WithInsecure(),
	)
	defer p.Shutdown(context.Background())

	// 初始化Tracer
	tracer, cfg := tracing.NewServerTracer()

	// 创建rpc客户端
	clients := client.InitClients()

	// 创建http处理器
	httpHandler := api_gateway2.NewHttpHandler(clients)

	// 启动hertz服务器
	h := server.Default(tracer)

	// 注册全局Tracing中间件
	h.Use(tracing.ServerMiddleware(cfg))

	// 注册路由
	router.InitRouters(h, httpHandler)

	h.Spin()
}
