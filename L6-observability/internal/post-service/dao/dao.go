package dao

import (
	"context"
	"errors"
	"log"
	"microservicesDemo/L6-observability/internal/post-service/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
	err := db.AutoMigrate(&model.Post{})
	if err != nil {
		log.Fatal("自动迁移Post模型失败")
		return nil
	}

	return &postRepositoryImpl{db: db}
}

func (p *postRepositoryImpl) AddPost(ctx context.Context, post *model.Post) error {
	tracer := otel.Tracer("post-service")
	ctx, span := tracer.Start(ctx, "dao.AddPost")
	defer span.End()

	span.SetAttributes(attribute.Bool("dao.success", true))

	return p.db.Create(post).Error
}

func (p *postRepositoryImpl) DeletePost(ctx context.Context, postID int64) error {
	tracer := otel.Tracer("post-service")
	ctx, span := tracer.Start(ctx, "dao.DeletePost")
	defer span.End()

	span.SetAttributes(attribute.Bool("dao.success", true))

	return p.db.Delete(&model.Post{}, postID).Error
}
