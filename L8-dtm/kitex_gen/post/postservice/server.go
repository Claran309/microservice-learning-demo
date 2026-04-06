package postservice

import (
	"context"

	"microservicesDemo/L8-dtm/kitex_gen/post"

	"github.com/cloudwego/kitex/server"
)

type PostService interface {
	CreatePost(ctx context.Context, req *post.CreatePostReq) (*post.CreatePostResp, error)
	DeletePost(ctx context.Context, req *post.DeletePostReq) (*post.DeletePostResp, error)
}

func NewServer(handler PostService, opts ...server.Option) server.Server {
	return server.NewServer(opts...)
}

type postServiceServer struct {
	handler PostService
}

func (s *postServiceServer) CreatePost(ctx context.Context, req *post.CreatePostReq) (*post.CreatePostResp, error) {
	return s.handler.CreatePost(ctx, req)
}

func (s *postServiceServer) DeletePost(ctx context.Context, req *post.DeletePostReq) (*post.DeletePostResp, error) {
	return s.handler.DeletePost(ctx, req)
}
