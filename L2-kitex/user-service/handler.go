package main

import (
	"context"
	"user-service/kitex_gen/user"
)

// UserServiceImpl 实现Kitex生成的接口
type UserServiceImpl struct{}

func (s *UserServiceImpl) Register(ctx context.Context, req *user.RegisterReq) (*user.RegisterResp, error) {
	if req.Username == "" || req.Password == "" {
		return &user.RegisterResp{
			Success: false,
			Message: "用户名、密码、邮箱不能为空",
		}, nil
	}

	if len(req.Username) < 3 {
		return &user.RegisterResp{
			Success: false,
			Message: "用户名至少3个字符",
		}, nil
	}

	return &user.RegisterResp{
		Success: true,
		Message: "注册成功",
	}, nil

}
