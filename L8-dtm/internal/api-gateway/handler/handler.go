package handler

import (
	"context"
	"encoding/json"
	"fmt"
	client2 "microservicesDemo/L8-dtm/internal/api-gateway/client"
	"microservicesDemo/L8-dtm/kitex_gen/post"
	"microservicesDemo/L8-dtm/kitex_gen/user"
	"microservicesDemo/L8-dtm/pkg/cache/redis"
	"microservicesDemo/L8-dtm/pkg/dtm"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
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

	zap.L().Info("√ 创建指标成功：requests_total，request_duration_seconds")
}

type HttpHandler struct {
	client     client2.Clients
	redisCache *redis.RedisCluster
	dtmManager *dtm.DTMManager
}

func NewHttpHandler(client client2.Clients, redisCache *redis.RedisCluster, dtmManager *dtm.DTMManager) *HttpHandler {
	initMetrics()

	return &HttpHandler{
		client:     client,
		redisCache: redisCache,
		dtmManager: dtmManager,
	}
}

func (h *HttpHandler) Register(c context.Context, ctx *app.RequestContext) {
	startTime := time.Now()

	tracer := otel.Tracer("api-gateway-handler")
	spanCtx, span := tracer.Start(c, "register-handler")
	defer span.End()
	zap.L().Info("√ 获取Trace成功：handler")

	span.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/api/v1/users/register"),
		attribute.String("user.agent", string(ctx.UserAgent())),
	)

	span.AddEvent("starting_user_registration")

	var req struct {
		Name     string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	if err := ctx.BindAndValidate(&req); err != nil {
		zap.L().Error("× 绑定参数时发生错误：",
			zap.Error(err),
			zap.String("handler", "register"),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(400, utils.H{
			"code": 400,
			"msg":  "绑定参数时发生错误",
		})
		return
	}

	resp, err := h.client.UserClient.Register(spanCtx, &user.RegisterReq{
		Username: req.Name,
		Password: req.Password,
		Email:    req.Email,
	})
	if err != nil {
		zap.L().Error("× 调用user-service失败：",
			zap.Error(err),
			zap.String("handler", "register"),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(500, utils.H{
			"code": 500,
			"msg":  "服务调用失败",
		})
		return
	}

	cacheKey := fmt.Sprintf("user:%d", resp.UserId)
	userData, _ := json.Marshal(map[string]interface{}{
		"user_id":  resp.UserId,
		"username": req.Name,
		"email":    req.Email,
	})
	h.redisCache.Set(spanCtx, cacheKey, string(userData), 30*time.Minute)

	span.SetAttributes(
		attribute.Int64("user.id", resp.UserId),
		attribute.Bool("handler.success", true),
	)

	duration := time.Since(startTime).Seconds()
	requestDuration.Record(spanCtx, duration)
	requestCounter.Add(spanCtx, 1)

	zap.L().Info("√ 用户注册成功",
		zap.Int64("user_id", resp.UserId),
		zap.String("username", req.Name),
	)

	ctx.JSON(200, utils.H{
		"code":    resp.Code,
		"msg":     resp.Msg,
		"user_id": resp.UserId,
	})
}

func (h *HttpHandler) Login(c context.Context, ctx *app.RequestContext) {
	startTime := time.Now()

	tracer := otel.Tracer("api-gateway-handler")
	spanCtx, span := tracer.Start(c, "login-handler")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/api/v1/users/login"),
	)

	var req struct {
		Name     string `json:"username"`
		Password string `json:"password"`
	}

	if err := ctx.BindAndValidate(&req); err != nil {
		zap.L().Error("× 绑定参数时发生错误：",
			zap.Error(err),
			zap.String("handler", "login"),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(400, utils.H{
			"code": 400,
			"msg":  "绑定参数时发生错误",
		})
		return
	}

	resp, err := h.client.UserClient.Login(spanCtx, &user.LoginReq{
		Username: req.Name,
		Password: req.Password,
	})
	if err != nil {
		zap.L().Error("× 调用user-service失败：",
			zap.Error(err),
			zap.String("handler", "login"),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(500, utils.H{
			"code": 500,
			"msg":  "服务调用失败",
		})
		return
	}

	span.SetAttributes(
		attribute.Int64("user.id", resp.UserId),
		attribute.Bool("handler.success", true),
	)

	duration := time.Since(startTime).Seconds()
	requestDuration.Record(spanCtx, duration)
	requestCounter.Add(spanCtx, 1)

	zap.L().Info("√ 用户登录成功",
		zap.Int64("user_id", resp.UserId),
	)

	ctx.JSON(200, utils.H{
		"code":    resp.Code,
		"msg":     resp.Msg,
		"user_id": resp.UserId,
	})
}

func (h *HttpHandler) CreatePost(c context.Context, ctx *app.RequestContext) {
	startTime := time.Now()

	tracer := otel.Tracer("api-gateway-handler")
	spanCtx, span := tracer.Start(c, "create-post-handler")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/api/v1/posts/create"),
	)

	var req struct {
		Owner   int64  `json:"owner"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := ctx.BindAndValidate(&req); err != nil {
		zap.L().Error("× 绑定参数时发生错误：",
			zap.Error(err),
			zap.String("handler", "create-post"),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(400, utils.H{
			"code": 400,
			"msg":  "绑定参数时发生错误",
		})
		return
	}

	resp, err := h.client.PostClient.CreatePost(spanCtx, &post.CreatePostReq{
		Owner:   req.Owner,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		zap.L().Error("× 调用post-service失败：",
			zap.Error(err),
			zap.String("handler", "create-post"),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(500, utils.H{
			"code": 500,
			"msg":  "服务调用失败",
		})
		return
	}

	span.SetAttributes(
		attribute.Int64("post.id", resp.PostId),
		attribute.Bool("handler.success", true),
	)

	duration := time.Since(startTime).Seconds()
	requestDuration.Record(spanCtx, duration)
	requestCounter.Add(spanCtx, 1)

	zap.L().Info("√ 创建文章成功",
		zap.Int64("post_id", resp.PostId),
	)

	ctx.JSON(200, utils.H{
		"code":    resp.Code,
		"msg":     resp.Msg,
		"post_id": resp.PostId,
	})
}

func (h *HttpHandler) CreatePostWithSaga(c context.Context, ctx *app.RequestContext) {
	startTime := time.Now()

	tracer := otel.Tracer("api-gateway-handler")
	spanCtx, span := tracer.Start(c, "create-post-saga-handler")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/api/v1/posts/saga/create"),
		attribute.String("transaction.type", "saga"),
	)

	var req struct {
		Owner   int64  `json:"owner"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := ctx.BindAndValidate(&req); err != nil {
		zap.L().Error("× 绑定参数时发生错误：",
			zap.Error(err),
			zap.String("handler", "create-post-saga"),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(400, utils.H{
			"code": 400,
			"msg":  "绑定参数时发生错误",
		})
		return
	}

	saga := h.dtmManager.NewSaga()
	gid := saga.GetGid()

	actionURL := dtm.BuildURL("localhost", 8888, "/api/v1/posts/create")
	compensateURL := dtm.BuildURL("localhost", 8888, "/api/v1/posts/compensate")

	payload := map[string]interface{}{
		"owner":   req.Owner,
		"title":   req.Title,
		"content": req.Content,
	}

	saga.AddStep(actionURL, compensateURL, payload)

	err := saga.Submit(spanCtx)
	if err != nil {
		zap.L().Error("× DTM SAGA事务提交失败：",
			zap.Error(err),
			zap.String("gid", gid),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(500, utils.H{
			"code": 500,
			"msg":  "分布式事务提交失败",
			"gid":  gid,
		})
		return
	}

	span.SetAttributes(
		attribute.String("dtm.gid", gid),
		attribute.Bool("handler.success", true),
	)

	duration := time.Since(startTime).Seconds()
	requestDuration.Record(spanCtx, duration)
	requestCounter.Add(spanCtx, 1)

	zap.L().Info("√ DTM SAGA事务提交成功",
		zap.String("gid", gid),
	)

	ctx.JSON(200, utils.H{
		"code": 200,
		"msg":  "分布式事务提交成功",
		"gid":  gid,
	})
}

func (h *HttpHandler) DeletePost(c context.Context, ctx *app.RequestContext) {
	startTime := time.Now()

	tracer := otel.Tracer("api-gateway-handler")
	spanCtx, span := tracer.Start(c, "delete-post-handler")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", "DELETE"),
		attribute.String("http.route", "/api/v1/posts/delete"),
	)

	postID := ctx.Query("post_id")
	if postID == "" {
		ctx.JSON(400, utils.H{
			"code": 400,
			"msg":  "post_id参数缺失",
		})
		return
	}

	var postIDInt int64
	fmt.Sscanf(postID, "%d", &postIDInt)

	resp, err := h.client.PostClient.DeletePost(spanCtx, &post.DeletePostReq{
		PostId: postIDInt,
	})
	if err != nil {
		zap.L().Error("× 调用post-service失败：",
			zap.Error(err),
			zap.String("handler", "delete-post"),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ctx.JSON(500, utils.H{
			"code": 500,
			"msg":  "服务调用失败",
		})
		return
	}

	span.SetAttributes(attribute.Bool("handler.success", true))

	duration := time.Since(startTime).Seconds()
	requestDuration.Record(spanCtx, duration)
	requestCounter.Add(spanCtx, 1)

	zap.L().Info("√ 删除文章成功",
		zap.String("post_id", postID),
	)

	ctx.JSON(200, utils.H{
		"code": resp.Code,
		"msg":  resp.Msg,
	})
}
