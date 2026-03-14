package service

import (
	"context"
	"errors"
	"fmt"
	"microservicesDemo/L5-kafka/internal/user-service/dao"
	"microservicesDemo/L5-kafka/internal/user-service/model"
	mq "microservicesDemo/L5-kafka/pkg/mq/kafka"
	"time"

	"github.com/cloudwego/hertz/pkg/common/json"
)

type UserService interface {
	Register(ctx context.Context, username string, email string, password string) (*model.User, error)
	Login(ctx context.Context, username string, password string) (*model.User, error)
}

type userServiceImpl struct {
	UserRepo      dao.UserRepository
	KafkaProducer *mq.Producer
}

func NewUserService(userRepo dao.UserRepository, kafkaProducer *mq.Producer) UserService {
	return &userServiceImpl{UserRepo: userRepo, KafkaProducer: kafkaProducer}
}

func (u *userServiceImpl) Register(ctx context.Context, username string, email string, password string) (*model.User, error) {
	if username == "" || email == "" || password == "" {
		return nil, errors.New("关键字段为空")
	}

	var user = &model.User{Username: username, Email: email, Password: password}

	err := u.UserRepo.AddUser(ctx, user)
	if err != nil {
		return nil, errors.New("添加用户失败: " + err.Error())
	}

	// 生产者发送消息
	if u.KafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"event_type":    "USER_REGISTER", // 明确的事件类型
			"user_id":       user.UserID,
			"username":      user.Username,
			"registered_at": time.Now(),
		})
		u.KafkaProducer.SendUserEvent(ctx, fmt.Sprintf("%d", user.UserID), eventData)
		fmt.Printf("发送用户注册事件:user_id:%d", user.UserID)
	}

	return user, nil
}

func (u *userServiceImpl) Login(ctx context.Context, username string, password string) (*model.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("关键字段为空")
	}

	user, err := u.UserRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("查找用户失败: " + err.Error())
	}

	if password != user.Password {
		return nil, errors.New("密码错误")
	}

	// 生产者发送消息
	if u.KafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"event_type":    "USER_LOGIN", // 明确的事件类型
			"user_id":       user.UserID,
			"username":      user.Username,
			"registered_at": time.Now(),
		})
		u.KafkaProducer.SendUserEvent(ctx, fmt.Sprintf("%d", user.UserID), eventData)
		fmt.Printf("发送用户登录事件:user_id:%d", user.UserID)
	}

	return user, nil
}
