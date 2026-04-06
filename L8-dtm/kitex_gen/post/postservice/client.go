package postservice

import (
	"context"

	"microservicesDemo/L8-dtm/kitex_gen/post"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/serviceinfo"
)

type Client interface {
	CreatePost(ctx context.Context, req *post.CreatePostReq) (*post.CreatePostResp, error)
	DeletePost(ctx context.Context, req *post.DeletePostReq) (*post.DeletePostResp, error)
}

type postServiceClient struct {
	client client.Client
}

func NewClient(destService string, opts ...client.Option) (Client, error) {
	svcInfo := &serviceinfo.ServiceInfo{
		ServiceName: "PostService",
		Methods:     map[string]serviceinfo.MethodInfo{},
	}
	c, err := client.NewClient(svcInfo, opts...)
	if err != nil {
		return nil, err
	}
	return &postServiceClient{client: c}, nil
}

func (c *postServiceClient) CreatePost(ctx context.Context, req *post.CreatePostReq) (*post.CreatePostResp, error) {
	resp := &post.CreatePostResp{}
	err := c.client.Call(ctx, "CreatePost", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *postServiceClient) DeletePost(ctx context.Context, req *post.DeletePostReq) (*post.DeletePostResp, error) {
	resp := &post.DeletePostResp{}
	err := c.client.Call(ctx, "DeletePost", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
