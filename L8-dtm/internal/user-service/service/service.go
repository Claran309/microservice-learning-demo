package service

import (
	"context"
	"errors"
	"microservicesDemo/L8-dtm/internal/user-service/dao"
	"microservicesDemo/L8-dtm/internal/user-service/model"
	mq "microservicesDemo/L8-dtm/pkg/mq/kafka"
	"time"

	"github.com/cloudwego/hertz/pkg/common/json"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type UserService interface {
	Register(ctx context.Context, username string, email string, password string) (*model.User, error)
	Login(ctx context.Context, username string, password string) (*model.User, error)
	GetUserByID(ctx context.Context, userID int64) (*model.User, error)
}

type userServiceImpl struct {
	UserRepo      dao.UserRepository
	KafkaProducer *mq.Producer
}

func NewUserService(userRepo dao.UserRepository, kafkaProducer *mq.Producer) UserService {
	zap.L().Info("√ 初始化UserService服务成功")
	return &userServiceImpl{UserRepo: userRepo, KafkaProducer: kafkaProducer}
}

func (u *userServiceImpl) Register(ctx context.Context, username string, email string, password string) (*model.User, error) {
	zap.L().Info("开始执行Register服务",
		zap.String("username", username),
		zap.String("email", email),
	)

	tracer := otel.Tracer("user-service")
	spanCtx, span := tracer.Start(ctx, "service.Register")
	defer span.End()

	if username == "" || email == "" || password == "" {
		zap.L().Error("× 关键字段为空",
			zap.String("service", "user-service"),
			zap.String("username", username),
			zap.String("email", email),
		)
		span.RecordError(errors.New("关键字段为空"))
		span.SetStatus(codes.Error, "关键字段为空")
		span.SetAttributes(attribute.Bool("service.success", false))
		return nil, errors.New("关键字段为空")
	}

	var user = &model.User{Username: username, Email: email, Password: password}
	err := u.UserRepo.AddUser(spanCtx, user)
	if err != nil {
		zap.L().Error("× 添加用户失败",
			zap.Error(err),
			zap.String("service", "user-service"),
			zap.String("username", username),
			zap.String("email", email),
		)
		span.RecordError(errors.New("添加用户失败: " + err.Error()))
		span.SetStatus(codes.Error, "添加用户失败: "+err.Error())
		span.SetAttributes(attribute.Bool("dao.success", false))
		return nil, errors.New("添加用户失败: " + err.Error())
	}
	zap.L().Info("√ 添加用户成功",
		zap.String("service", "user-service"),
		zap.Int64("user_id", user.UserID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)

	span.SetAttributes(
		attribute.String("username", user.Username),
		attribute.String("email", user.Email),
		attribute.Bool("dao.success", true),
	)

	if u.KafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"event_type":    "USER_REGISTER",
			"user_id":       user.UserID,
			"username":      user.Username,
			"registered_at": time.Now(),
		})

		err := u.KafkaProducer.SendUserEvent(ctx, user.Username, eventData)
		if err != nil {
			zap.L().Error("× 发送用户注册事件失败",
				zap.Error(err),
				zap.String("service", "user-service"),
				zap.Int64("user_id", user.UserID),
			)
		} else {
			zap.L().Info("√ 发送用户注册事件成功",
				zap.Int64("user_id", user.UserID),
			)
		}
	}

	span.SetAttributes(attribute.Bool("service.success", true))

	return user, nil
}

func (u *userServiceImpl) Login(ctx context.Context, username string, password string) (*model.User, error) {
	zap.L().Info("开始执行Login服务",
		zap.String("username", username),
	)

	tracer := otel.Tracer("user-service")
	spanCtx, span := tracer.Start(ctx, "service.Login")
	defer span.End()

	user, err := u.UserRepo.GetUserByUsername(spanCtx, username)
	if err != nil {
		zap.L().Error("× 用户不存在",
			zap.Error(err),
			zap.String("username", username),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "用户不存在")
		return nil, errors.New("用户不存在")
	}

	if user.Password != password {
		zap.L().Error("× 密码错误",
			zap.String("username", username),
		)
		span.RecordError(errors.New("密码错误"))
		span.SetStatus(codes.Error, "密码错误")
		return nil, errors.New("密码错误")
	}

	span.SetAttributes(
		attribute.Int64("user.id", user.UserID),
		attribute.Bool("service.success", true),
	)

	zap.L().Info("√ 用户登录成功",
		zap.Int64("user_id", user.UserID),
		zap.String("username", username),
	)

	return user, nil
}

func (u *userServiceImpl) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	zap.L().Info("开始执行GetUserByID服务",
		zap.Int64("user_id", userID),
	)

	tracer := otel.Tracer("user-service")
	spanCtx, span := tracer.Start(ctx, "service.GetUserByID")
	defer span.End()

	user, err := u.UserRepo.GetUserByID(spanCtx, userID)
	if err != nil {
		zap.L().Error("× 获取用户失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Bool("service.success", true))

	return user, nil
}
