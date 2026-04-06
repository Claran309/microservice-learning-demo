package handler

import (
	"context"
	"microservicesDemo/L8-dtm/internal/user-service/service"
	"microservicesDemo/L8-dtm/kitex_gen/user"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

var (
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	errorCounter    metric.Int64Counter
)

func initMetrics() {
	meter := otel.Meter("user-service.handler")

	var err error

	requestCounter, err = meter.Int64Counter(
		"claran_user_request_total",
		metric.WithDescription("总请求数"),
	)
	if err != nil {
		zap.L().Fatal("× 创建requestCounter失败",
			zap.Error(err),
		)
	}

	requestDuration, err = meter.Float64Histogram(
		"claran_user_request_duration_seconds",
		metric.WithDescription("请求耗时"),
		metric.WithUnit("s"),
	)
	if err != nil {
		zap.L().Fatal("× 创建requestDuration失败",
			zap.Error(err),
		)
	}

	errorCounter, err = meter.Int64Counter(
		"claran_user_errors_total",
		metric.WithDescription("总错误数"),
	)
	if err != nil {
		zap.L().Fatal("× 创建errorCounter失败",
			zap.Error(err),
		)
	}
}

type UserServiceImpl struct {
	UserService service.UserService
}

func NewUserServiceImpl(userService service.UserService) *UserServiceImpl {
	initMetrics()
	zap.L().Info("√ 初始化UserService处理器成功")
	return &UserServiceImpl{UserService: userService}
}

func (s *UserServiceImpl) Register(ctx context.Context, req *user.RegisterReq) (resp *user.RegisterResp, err error) {
	startTime := time.Now()
	zap.L().Info("开始处理Register请求",
		zap.String("username", req.Username),
		zap.String("email", req.Email),
	)

	defer func() {
		duration := time.Since(startTime).Seconds()
		status := getstatus(err)

		requestDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("method", "Register"),
				attribute.String("status", status),
			))

		requestCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", "Register"),
				attribute.String("status", status),
			))

		if err != nil {
			errorCounter.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("method", "Register"),
				))
		}
	}()

	tracer := otel.Tracer("user-service-handler")
	ctx, span := tracer.Start(ctx, "handler.Register")
	defer span.End()

	userModel, err := s.UserService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		zap.L().Error("× Register处理失败",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return &user.RegisterResp{
			Code: 400,
			Msg:  err.Error(),
		}, nil
	}

	span.SetAttributes(
		attribute.Int64("user.id", userModel.UserID),
		attribute.Bool("handler.success", true),
	)

	zap.L().Info("√ Register处理成功",
		zap.Int64("user_id", userModel.UserID),
	)

	return &user.RegisterResp{
		Code:    200,
		Msg:     "注册成功",
		UserId:  userModel.UserID,
		Success: true,
	}, nil
}

func (s *UserServiceImpl) Login(ctx context.Context, req *user.LoginReq) (resp *user.LoginResp, err error) {
	startTime := time.Now()
	zap.L().Info("开始处理Login请求",
		zap.String("username", req.Username),
	)

	defer func() {
		duration := time.Since(startTime).Seconds()
		status := getstatus(err)

		requestDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("method", "Login"),
				attribute.String("status", status),
			))

		requestCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", "Login"),
				attribute.String("status", status),
			))
	}()

	tracer := otel.Tracer("user-service-handler")
	ctx, span := tracer.Start(ctx, "handler.Login")
	defer span.End()

	userModel, err := s.UserService.Login(ctx, req.Username, req.Password)
	if err != nil {
		zap.L().Error("× Login处理失败",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return &user.LoginResp{
			Code: 400,
			Msg:  err.Error(),
		}, nil
	}

	span.SetAttributes(
		attribute.Int64("user.id", userModel.UserID),
		attribute.Bool("handler.success", true),
	)

	zap.L().Info("√ Login处理成功",
		zap.Int64("user_id", userModel.UserID),
	)

	return &user.LoginResp{
		Code:    200,
		Msg:     "登录成功",
		UserId:  userModel.UserID,
		Success: true,
	}, nil
}

func getstatus(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}
