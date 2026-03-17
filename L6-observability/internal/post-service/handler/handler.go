package handler

import (
	"context"
	"fmt"
	"microservicesDemo/L6-observability/internal/post-service/service"
	"microservicesDemo/L6-observability/kitex_gen/post"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	// Metrics指标
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	errorCounter    metric.Int64Counter
)

func initMetrics() {
	// 从全局获取 meter
	meter := otel.Meter("post-service.handler")

	var err error

	requestCounter, err = meter.Int64Counter(
		"claran_post_request_total",
		metric.WithDescription("总请求数"),
	)
	if err != nil {
		panic(fmt.Sprintf("创建requestCounter失败: %v", err))
	}

	requestDuration, err = meter.Float64Histogram(
		"claran_post_request_duration_seconds",
		metric.WithDescription("请求耗时"),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(fmt.Sprintf("创建requestDuration失败: %v", err))
	}

	errorCounter, err = meter.Int64Counter(
		"claran_post_errors_total",
		metric.WithDescription("总错误数"),
	)
	if err != nil {
		panic(fmt.Sprintf("创建errorCounter失败: %v", err))
	}
}

// PostServiceImpl implements the last service interface defined in the IDL.
type PostServiceImpl struct {
	PostService service.PostService
}

func NewPostServiceImpl(postService service.PostService) *PostServiceImpl {
	initMetrics()
	return &PostServiceImpl{PostService: postService}
}

// CreatePost implements the PostServiceImpl interface.
func (s *PostServiceImpl) CreatePost(ctx context.Context, req *post.CreatePostReq) (resp *post.CreatePostResp, err error) {
	startTime := time.Now()

	// 记录指标
	defer func() {
		duration := time.Since(startTime).Seconds()
		requestDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("method", "Register"),
				attribute.String("status", getstatus(err)),
			))

		requestCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", "Register"),
				attribute.String("status", getstatus(err)),
			))

		if err != nil {
			errorCounter.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("method", "Register"),
					attribute.String("status", getstatus(err)),
				))
		}
	}()

	// 创建span - 从全局获取 tracer
	tracer := otel.Tracer("post-service")
	spanCtx, span := tracer.Start(ctx, "handler.CreatePost",
		trace.WithAttributes(
			attribute.String("post_name", req.PostName),
		))
	defer span.End()

	posts, err := s.PostService.CreatePost(spanCtx, req.PostName, req.UserID, req.Content)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "执行创建文章服务失败："+err.Error())
		span.SetAttributes(attribute.Bool("service.success", false))
		return nil, err
	}

	span.SetAttributes(
		attribute.String("post_title", posts.Title),
		attribute.String("post_id", strconv.Itoa(int(posts.PostID))),
		attribute.String("content", posts.Content),
		attribute.String("owner_id", strconv.Itoa(int(posts.Owner))),
		attribute.Bool("service.success", true),
	)

	resp = &post.CreatePostResp{
		Success: true,
		PostID:  posts.PostID, // 乱写的，前期没构建好导致的
		Msg:     "success",
	}

	return resp, nil
}

// DeletePost implements the PostServiceImpl interface.
func (s *PostServiceImpl) DeletePost(ctx context.Context, req *post.DeletePostReq) (resp *post.DeletePostResp, err error) {
	startTime := time.Now()

	// 记录指标
	defer func() {
		duration := time.Since(startTime).Seconds()
		requestDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("method", "Register"),
				attribute.String("status", getstatus(err)),
			))

		requestCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", "Register"),
				attribute.String("status", getstatus(err)),
			))

		if err != nil {
			errorCounter.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("method", "Register"),
					attribute.String("status", getstatus(err)),
				))
		}
	}()

	// 创建span - 从全局获取 tracer
	tracer := otel.Tracer("post-service")
	spanCtx, span := tracer.Start(ctx, "handler.DeletePost",
		trace.WithAttributes(
			attribute.String("post_id", strconv.Itoa(int(req.PostID))),
		))
	defer span.End()

	err = s.PostService.DeletePost(spanCtx, req.PostID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "执行删除文章服务失败："+err.Error())
		span.SetAttributes(attribute.Bool("service.success", false))
		return nil, err
	}

	resp = &post.DeletePostResp{
		Success: true,
		Msg:     "success",
	}

	span.SetAttributes(
		attribute.Bool("service.success", true),
	)

	return resp, nil
}

func getstatus(err error) string {
	if err == nil {
		return "success"
	}
	return "error"
}
