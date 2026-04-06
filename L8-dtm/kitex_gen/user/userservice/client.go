package userservice

import (
	"context"

	"microservicesDemo/L8-dtm/kitex_gen/user"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/serviceinfo"
)

type Client interface {
	Register(ctx context.Context, req *user.RegisterReq) (*user.RegisterResp, error)
	Login(ctx context.Context, req *user.LoginReq) (*user.LoginResp, error)
}

type userServiceClient struct {
	client client.Client
}

func NewClient(destService string, opts ...client.Option) (Client, error) {
	svcInfo := &serviceinfo.ServiceInfo{
		ServiceName: "UserService",
		Methods:     map[string]serviceinfo.MethodInfo{},
	}
	c, err := client.NewClient(svcInfo, opts...)
	if err != nil {
		return nil, err
	}
	return &userServiceClient{client: c}, nil
}

func (c *userServiceClient) Register(ctx context.Context, req *user.RegisterReq) (*user.RegisterResp, error) {
	resp := &user.RegisterResp{}
	err := c.client.Call(ctx, "Register", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *userServiceClient) Login(ctx context.Context, req *user.LoginReq) (*user.LoginResp, error) {
	resp := &user.LoginResp{}
	err := c.client.Call(ctx, "Login", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
