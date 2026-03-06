package service

import (
	"context"
	"errors"
	"microservicesDemo/L4-govern/internal/user-service/dao"
	"microservicesDemo/L4-govern/internal/user-service/model"
)

type UserService interface {
	Register(ctx context.Context, username string, email string, password string) (*model.User, error)
	Login(ctx context.Context, username string, password string) (*model.User, error)
}

type userServiceImpl struct {
	UserRepo dao.UserRepository
}

func NewUserService(userRepo dao.UserRepository) UserService {
	return &userServiceImpl{UserRepo: userRepo}
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

	return user, nil
}
