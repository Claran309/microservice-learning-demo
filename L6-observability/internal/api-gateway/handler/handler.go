package handler

import (
	"context"
	"log"
	client2 "microservicesDemo/L6-observability/internal/api-gateway/client"
	"microservicesDemo/L6-observability/kitex_gen/post"
	"microservicesDemo/L6-observability/kitex_gen/user"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

var (
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
)

func initMetrics() {
	meter := otel.GetMeterProvider().Meter("api-gateway")

	var err error
	requestCounter, err = meter.Int64Counter(
		"claran_api_gateway_requests_total",
		metric.WithDescription("Total number of API requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		hlog.Errorf("Failed to create request counter: %v", err)
	}

	requestDuration, err = meter.Float64Histogram(
		"claran_api_gateway_request_duration_seconds",
		metric.WithDescription("API request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		hlog.Errorf("Failed to create request duration histogram: %v", err)
	}
}

type HttpHandler struct {
	client client2.Clients
}

func NewHttpHandler(client client2.Clients) *HttpHandler {
	initMetrics()

	return &HttpHandler{
		client: client,
	}
}

func (h *HttpHandler) Register(c context.Context, ctx *app.RequestContext) {
	startTime := time.Now()

	// 获取当前Trace上下文
	tracer := otel.Tracer("api-gateway-handler")
	spanCtx, span := tracer.Start(c, "register-handler")
	defer span.End()

	// 设置Span属性
	span.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/api/v1/users/register"),
		attribute.String("user.agent", string(ctx.UserAgent())),
	)

	// 记录自定义事件
	span.AddEvent("starting_user_registration")

	var req struct {
		Name     string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	if err := ctx.BindAndValidate(&req); err != nil {
		log.Println("绑定参数时发生错误", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(400, utils.H{
			"code": 400,
			"msg":  "绑定参数时发生错误",
		})
		return
	}

	rpcReq := user.RegisterReq{
		Username: req.Name,
		Password: req.Password,
		Email:    req.Email,
	}

	rpcResp, err := h.client.UserClient.Register(c, &rpcReq)
	if err != nil {
		log.Printf("RPC调用失败: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(500, utils.H{
			"code":    500,
			"message": "服务内部错误: " + err.Error(),
		})
		return
	}

	statusCode := 200
	if !rpcResp.Success {
		statusCode = 400
	}
	duration := time.Since(startTime)
	durationSeconds := time.Since(startTime).Seconds()

	requestCounter.Add(spanCtx, 1,
		metric.WithAttributes(
			attribute.String("endpoint", "/register"),
			attribute.String("method", "POST"),
			attribute.String("status", "success"),
		))
	requestDuration.Record(spanCtx, durationSeconds,
		metric.WithAttributes(
			attribute.String("endpoint", "/register"),
		))

	// 返回HTTP响应
	log.Printf("注册处理完成，状态码: %d，耗时: %v", statusCode, duration)

	ctx.JSON(statusCode, utils.H{
		"code":    statusCode,      // HTTP状态码
		"success": rpcResp.Success, // 业务是否成功
		"message": rpcResp.Msg,     // 业务消息
	})
}

func (h *HttpHandler) Login(c context.Context, ctx *app.RequestContext) {
	startTime := time.Now()

	// 获取当前Trace上下文
	tracer := otel.Tracer("api-gateway-handler")
	spanCtx, span := tracer.Start(c, "login-handler")
	defer span.End()

	// 设置Span属性
	span.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/api/v1/users/login"),
		attribute.String("user.agent", string(ctx.UserAgent())),
	)

	// 记录自定义事件
	span.AddEvent("starting_user_login")

	var req struct {
		Name     string `json:"username"`
		Password string `json:"password"`
	}

	if err := ctx.BindAndValidate(&req); err != nil {
		log.Println("绑定参数时发生错误", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(400, utils.H{
			"code": 400,
			"msg":  "绑定参数时发生错误",
		})
		return
	}

	rpcReq := user.LoginByUsernameReq{
		Username: req.Name,
		Password: req.Password,
	}

	rpcResp, err := h.client.UserClient.Login(c, &rpcReq)
	if err != nil {
		log.Printf("RPC调用失败: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(500, utils.H{
			"code":    500,
			"message": "服务内部错误: " + err.Error(),
		})
		return
	}

	statusCode := 200
	if !rpcResp.Success {
		statusCode = 400
	}
	duration := time.Since(startTime)
	durationSeconds := time.Since(startTime).Seconds()

	requestCounter.Add(spanCtx, 1,
		metric.WithAttributes(
			attribute.String("endpoint", "/login"),
			attribute.String("method", "POST"),
			attribute.String("status", "success"),
		))
	requestDuration.Record(spanCtx, durationSeconds,
		metric.WithAttributes(
			attribute.String("endpoint", "/login"),
		))

	// 返回HTTP响应
	log.Printf("注册处理完成，状态码: %d，耗时: %v", statusCode, duration)

	ctx.JSON(statusCode, utils.H{
		"code":    statusCode,      // HTTP状态码
		"success": rpcResp.Success, // 业务是否成功
		"message": rpcResp.Msg,     // 业务消息
	})
}

func (h *HttpHandler) CreatePost(c context.Context, ctx *app.RequestContext) {
	startTime := time.Now()

	var req struct {
		UserID  int64  `json:"user_id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	rpcReq := post.CreatePostReq{
		UserID:   req.UserID,
		PostName: req.Title,
		Content:  req.Content,
	}

	rpcResp, err := h.client.PostClient.CreatePost(c, &rpcReq)
	if err != nil {
		log.Printf("RPC调用失败: %v", err)
		ctx.JSON(500, utils.H{
			"code":    500,
			"message": "服务内部错误: " + err.Error(),
		})
		return
	}

	statusCode := 200
	if !rpcResp.Success {
		statusCode = 400
	}
	duration := time.Since(startTime)

	// 返回HTTP响应
	log.Printf("注册处理完成，状态码: %d，耗时: %v", statusCode, duration)

	ctx.JSON(statusCode, utils.H{
		"code":    statusCode,      // HTTP状态码
		"success": rpcResp.Success, // 业务是否成功
		"message": rpcResp.Msg,     // 业务消息
	})
}

func (h *HttpHandler) DeletePost(c context.Context, ctx *app.RequestContext) {}
