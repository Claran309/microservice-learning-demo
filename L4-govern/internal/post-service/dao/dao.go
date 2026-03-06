package dao

import (
	"context"
	"errors"
	"microservicesDemo/L4-govern/internal/post-service/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		return nil, errors.New("数据库连接失败：" + err.Error())
	}
	return db, nil
}

type PostRepository interface {
	AddPost(ctx context.Context, post *model.Post) error
	DeletePost(ctx context.Context, postID int64) error
}

type postRepositoryImpl struct {
	db *gorm.DB
}

func NewPostRepositoryImpl(db *gorm.DB) PostRepository {
	return &postRepositoryImpl{db: db}
}

func (p *postRepositoryImpl) AddPost(ctx context.Context, post *model.Post) error {
	return p.db.Create(post).Error
}

func (p *postRepositoryImpl) DeletePost(ctx context.Context, postID int64) error {
	return p.db.Delete(&model.Post{}, postID).Error
}
