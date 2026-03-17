package handler

import (
	"context"
	"fmt"
	"log"
	"microservicesDemo/L6-observability/internal/user-service/service"
	"microservicesDemo/L6-observability/kitex_gen/user"
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
	meter := otel.Meter("user-service.handler")

	var err error

	requestCounter, err = meter.Int64Counter(
		"claran_user_request_total",
		metric.WithDescription("总请求数"),
	)
	if err != nil {
		panic(fmt.Sprintf("创建requestCounter失败: %v", err))
	}

	requestDuration, err = meter.Float64Histogram(
		"claran_user_request_duration_seconds",
		metric.WithDescription("请求耗时"),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(fmt.Sprintf("创建requestDuration失败: %v", err))
	}

	errorCounter, err = meter.Int64Counter(
		"claran_user_errors_total",
		metric.WithDescription("总错误数"),
	)
	if err != nil {
		panic(fmt.Sprintf("创建errorCounter失败: %v", err))
	}
}

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct {
	UserService service.UserService
}

func NewUserServiceImpl(userService service.UserService) *UserServiceImpl {
	initMetrics()
	return &UserServiceImpl{UserService: userService}
}

// Register implements the UserServiceImpl interface.
func (s *UserServiceImpl) Register(ctx context.Context, req *user.RegisterReq) (resp *user.RegisterResp, err error) {
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
	tracer := otel.Tracer("user-service")
	spanCtx, span := tracer.Start(ctx, "handler.Register",
		trace.WithAttributes(
			attribute.String("username", req.Username),
		))
	defer span.End()

	//sCtx := trace.SpanContextFromContext(ctx)
	//log.Printf("Handler -> Service, TraceID: %v", sCtx.TraceID())
	newUser, err := s.UserService.Register(spanCtx, req.Username, req.Email, req.Password)
	if err != nil {
		log.Println("执行注册服务失败：" + err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, "执行注册服务失败："+err.Error())
		span.SetAttributes(attribute.Bool("service.success", false))
		return resp, err
	}

	span.SetAttributes(
		attribute.String("username", newUser.Username),
		attribute.String("email", newUser.Email),
		attribute.Bool("service.success", true),
	)

	resp = &user.RegisterResp{
		Success: true,
		UserID:  newUser.UserID,
		Msg:     "注册成功",
	}

	return resp, nil
}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, req *user.LoginByUsernameReq) (resp *user.LoginByUsernameResp, err error) {
	startTime := time.Now()

	// 记录指标
	defer func() {
		duration := time.Since(startTime).Seconds()
		requestDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("method", "Login"),
				attribute.String("status", getstatus(err)),
			))

		requestCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", "Login"),
				attribute.String("status", getstatus(err)),
			))

		if err != nil {
			errorCounter.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("method", "Login"),
					attribute.String("status", getstatus(err)),
				))
		}
	}()

	// 创建span - 从全局获取 tracer
	tracer := otel.Tracer("user-service")
	spanCtx, span := tracer.Start(ctx, "handler.Login",
		trace.WithAttributes(
			attribute.String("username", req.Username),
		))
	defer span.End()

	newUser, err := s.UserService.Login(spanCtx, req.Username, req.Password)
	if err != nil {
		log.Println("执行登录服务失败：" + err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, "执行登录服务失败："+err.Error())
		span.SetAttributes(attribute.Bool("service.success", false))
		return resp, err
	}

	span.SetAttributes(
		attribute.String("username", newUser.Username),
		attribute.String("email", newUser.Email),
		attribute.Bool("service.success", true),
	)

	resp = &user.LoginByUsernameResp{
		Success: true,
		Token:   newUser.Username,
		Msg:     "登录成功",
	}

	return resp, err
}

func getstatus(err error) string {
	if err == nil {
		return "success"
	}
	return "error"
}
