package userservice

import (
	"context"

	"microservicesDemo/L8-dtm/kitex_gen/user"

	"github.com/cloudwego/kitex/server"
)

type UserService interface {
	Register(ctx context.Context, req *user.RegisterReq) (*user.RegisterResp, error)
	Login(ctx context.Context, req *user.LoginReq) (*user.LoginResp, error)
}

func NewServer(handler UserService, opts ...server.Option) server.Server {
	return server.NewServer(opts...)
}

type userServiceServer struct {
	handler UserService
}

func (s *userServiceServer) Register(ctx context.Context, req *user.RegisterReq) (*user.RegisterResp, error) {
	return s.handler.Register(ctx, req)
}

func (s *userServiceServer) Login(ctx context.Context, req *user.LoginReq) (*user.LoginResp, error) {
	return s.handler.Login(ctx, req)
}
