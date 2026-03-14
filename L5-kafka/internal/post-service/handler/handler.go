package handler

import (
	"context"
	"microservicesDemo/L5-kafka/internal/post-service/service"
	"microservicesDemo/L5-kafka/kitex_gen/post"
)

// PostServiceImpl implements the last service interface defined in the IDL.
type PostServiceImpl struct {
	PostService service.PostService
}

func NewPostServiceImpl(postService service.PostService) *PostServiceImpl {
	return &PostServiceImpl{PostService: postService}
}

// CreatePost implements the PostServiceImpl interface.
func (s *PostServiceImpl) CreatePost(ctx context.Context, req *post.CreatePostReq) (resp *post.CreatePostResp, err error) {
	posts, err := s.PostService.CreatePost(ctx, req.PostName, req.UserID, req.Content)
	if err != nil {
		return nil, err
	}

	resp = &post.CreatePostResp{
		Success: true,
		PostID:  posts.PostID, // 乱写的，前期没构建好导致的
		Msg:     "success",
	}

	return resp, nil
}

// DeletePost implements the PostServiceImpl interface.
func (s *PostServiceImpl) DeletePost(ctx context.Context, req *post.DeletePostReq) (resp *post.DeletePostResp, err error) {
	// TODO: Your code here...
	err = s.PostService.DeletePost(ctx, req.PostID)
	if err != nil {
		return nil, err
	}

	resp = &post.DeletePostResp{
		Success: true,
		Msg:     "success",
	}

	return resp, nil
}
