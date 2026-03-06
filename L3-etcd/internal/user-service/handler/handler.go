package handler

import (
	"context"
	"log"
	"microservicesDemo/L3-etcd/internal/user-service/service"
	"microservicesDemo/L3-etcd/kitex_gen/user"
)

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct {
	UserService service.UserService
}

func NewUserServiceImpl(userService service.UserService) *UserServiceImpl {
	return &UserServiceImpl{UserService: userService}
}

// Register implements the UserServiceImpl interface.
func (s *UserServiceImpl) Register(ctx context.Context, req *user.RegisterReq) (resp *user.RegisterResp, err error) {
	newUser, err := s.UserService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		log.Println("执行注册服务失败：" + err.Error())
		return resp, err
	}

	resp = &user.RegisterResp{
		Success: true,
		UserID:  newUser.UserID,
		Msg:     "注册成功",
	}

	return resp, nil
}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, req *user.LoginByUsernameReq) (resp *user.LoginByUsernameResp, err error) {
	newUser, err := s.UserService.Login(ctx, req.Username, req.Password)
	if err != nil {
		log.Println("执行登录服务失败：" + err.Error())
		return resp, err
	}

	resp = &user.LoginByUsernameResp{
		Success: true,
		Token:   newUser.Username,
		Msg:     "登录成功",
	}

	return resp, err
}
