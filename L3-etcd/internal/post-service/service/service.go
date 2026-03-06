package service

import (
	"context"
	"microservicesDemo/L3-etcd/internal/post-service/dao"
	"microservicesDemo/L3-etcd/internal/post-service/model"
)

type PostService interface {
	CreatePost(ctx context.Context, title string, userID int64, content string) (*model.Post, error)
	DeletePost(ctx context.Context, postID int64) error
}

type postServiceImpl struct {
	PostRepo dao.PostRepository
}

func NewPostServiceImpl(PostRepo dao.PostRepository) PostService {
	return &postServiceImpl{PostRepo: PostRepo}
}

func (s *postServiceImpl) CreatePost(ctx context.Context, title string, userID int64, content string) (*model.Post, error) {
	var post = model.Post{
		Title:   title,
		Content: content,
		Owner:   userID,
	}

	err := s.PostRepo.AddPost(ctx, &post)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *postServiceImpl) DeletePost(ctx context.Context, postID int64) error {
	return s.PostRepo.DeletePost(ctx, postID)
}
