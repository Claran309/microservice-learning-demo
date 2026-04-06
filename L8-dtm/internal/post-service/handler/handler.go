package handler

import (
	"context"
	"microservicesDemo/L8-dtm/internal/post-service/service"
	"microservicesDemo/L8-dtm/kitex_gen/post"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

var (
	postRequestCounter  metric.Int64Counter
	postRequestDuration metric.Float64Histogram
	postErrorCounter    metric.Int64Counter
)

func initPostMetrics() {
	meter := otel.Meter("post-service.handler")

	var err error

	postRequestCounter, err = meter.Int64Counter(
		"claran_post_request_total",
		metric.WithDescription("总请求数"),
	)
	if err != nil {
		zap.L().Fatal("× 创建postRequestCounter失败", zap.Error(err))
	}

	postRequestDuration, err = meter.Float64Histogram(
		"claran_post_request_duration_seconds",
		metric.WithDescription("请求耗时"),
		metric.WithUnit("s"),
	)
	if err != nil {
		zap.L().Fatal("× 创建postRequestDuration失败", zap.Error(err))
	}

	postErrorCounter, err = meter.Int64Counter(
		"claran_post_errors_total",
		metric.WithDescription("总错误数"),
	)
	if err != nil {
		zap.L().Fatal("× 创建postErrorCounter失败", zap.Error(err))
	}
}

type PostServiceImpl struct {
	PostService service.PostService
}

func NewPostServiceImpl(postService service.PostService) *PostServiceImpl {
	initPostMetrics()
	zap.L().Info("√ 初始化PostService处理器成功")
	return &PostServiceImpl{PostService: postService}
}

func (p *PostServiceImpl) CreatePost(ctx context.Context, req *post.CreatePostReq) (resp *post.CreatePostResp, err error) {
	startTime := time.Now()
	zap.L().Info("开始处理CreatePost请求",
		zap.Int64("owner", req.Owner),
		zap.String("title", req.Title),
	)

	defer func() {
		duration := time.Since(startTime).Seconds()
		status := getPostStatus(err)

		postRequestDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("method", "CreatePost"),
				attribute.String("status", status),
			))

		postRequestCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", "CreatePost"),
				attribute.String("status", status),
			))

		if err != nil {
			postErrorCounter.Add(ctx, 1,
				metric.WithAttributes(attribute.String("method", "CreatePost")))
		}
	}()

	tracer := otel.Tracer("post-service-handler")
	ctx, span := tracer.Start(ctx, "handler.CreatePost")
	defer span.End()

	postModel, err := p.PostService.CreatePost(ctx, req.Owner, req.Title, req.Content)
	if err != nil {
		zap.L().Error("× CreatePost处理失败",
			zap.Error(err),
			zap.Int64("owner", req.Owner),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return &post.CreatePostResp{
			Code: 400,
			Msg:  err.Error(),
		}, nil
	}

	span.SetAttributes(
		attribute.Int64("post.id", postModel.PostID),
		attribute.Bool("handler.success", true),
	)

	zap.L().Info("√ CreatePost处理成功",
		zap.Int64("post_id", postModel.PostID),
	)

	return &post.CreatePostResp{
		Code:    200,
		Msg:     "创建文章成功",
		PostId:  postModel.PostID,
		Success: true,
	}, nil
}

func (p *PostServiceImpl) DeletePost(ctx context.Context, req *post.DeletePostReq) (resp *post.DeletePostResp, err error) {
	startTime := time.Now()
	zap.L().Info("开始处理DeletePost请求",
		zap.Int64("post_id", req.PostId),
	)

	defer func() {
		duration := time.Since(startTime).Seconds()
		status := getPostStatus(err)

		postRequestDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("method", "DeletePost"),
				attribute.String("status", status),
			))

		postRequestCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", "DeletePost"),
				attribute.String("status", status),
			))
	}()

	tracer := otel.Tracer("post-service-handler")
	ctx, span := tracer.Start(ctx, "handler.DeletePost")
	defer span.End()

	err = p.PostService.DeletePost(ctx, req.PostId)
	if err != nil {
		zap.L().Error("× DeletePost处理失败",
			zap.Error(err),
			zap.Int64("post_id", req.PostId),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return &post.DeletePostResp{
			Code: 400,
			Msg:  err.Error(),
		}, nil
	}

	span.SetAttributes(attribute.Bool("handler.success", true))

	zap.L().Info("√ DeletePost处理成功",
		zap.Int64("post_id", req.PostId),
	)

	return &post.DeletePostResp{
		Code:    200,
		Msg:     "删除文章成功",
		Success: true,
	}, nil
}

func getPostStatus(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}
