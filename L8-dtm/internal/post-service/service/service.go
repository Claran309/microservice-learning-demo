package service

import (
	"context"
	"errors"
	"microservicesDemo/L8-dtm/internal/post-service/dao"
	"microservicesDemo/L8-dtm/internal/post-service/model"
	mq "microservicesDemo/L8-dtm/pkg/mq/kafka"
	"time"

	"github.com/cloudwego/hertz/pkg/common/json"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type PostService interface {
	CreatePost(ctx context.Context, owner int64, title string, content string) (*model.Post, error)
	DeletePost(ctx context.Context, postID int64) error
	GetPostByID(ctx context.Context, postID int64) (*model.Post, error)
	GetPostsByOwner(ctx context.Context, ownerID int64) ([]model.Post, error)
}

type postServiceImpl struct {
	PostRepo     dao.PostRepository
	KafkaProducer *mq.Producer
}

func NewPostService(postRepo dao.PostRepository, kafkaProducer *mq.Producer) PostService {
	zap.L().Info("√ 初始化PostService服务成功")
	return &postServiceImpl{PostRepo: postRepo, KafkaProducer: kafkaProducer}
}

func (p *postServiceImpl) CreatePost(ctx context.Context, owner int64, title string, content string) (*model.Post, error) {
	zap.L().Info("开始执行CreatePost服务",
		zap.Int64("owner", owner),
		zap.String("title", title),
	)

	tracer := otel.Tracer("post-service")
	spanCtx, span := tracer.Start(ctx, "service.CreatePost")
	defer span.End()

	if title == "" || content == "" {
		zap.L().Error("× 关键字段为空",
			zap.String("service", "post-service"),
		)
		span.RecordError(errors.New("关键字段为空"))
		span.SetStatus(codes.Error, "关键字段为空")
		return nil, errors.New("标题和内容不能为空")
	}

	post := &model.Post{
		Owner:   owner,
		Title:   title,
		Content: content,
	}

	err := p.PostRepo.AddPost(spanCtx, post)
	if err != nil {
		zap.L().Error("× 创建文章失败",
			zap.Error(err),
			zap.Int64("owner", owner),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(
		attribute.Int64("post.id", post.PostID),
		attribute.Bool("service.success", true),
	)

	if p.KafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"event_type":   "POST_CREATE",
			"post_id":      post.PostID,
			"owner_id":     post.Owner,
			"created_at":   time.Now(),
		})

		err := p.KafkaProducer.SendUserEvent(ctx, string(post.PostID), eventData)
		if err != nil {
			zap.L().Error("× 发送文章创建事件失败",
				zap.Error(err),
				zap.Int64("post_id", post.PostID),
			)
		}
	}

	zap.L().Info("√ 创建文章成功",
		zap.Int64("post_id", post.PostID),
		zap.Int64("owner", owner),
	)

	return post, nil
}

func (p *postServiceImpl) DeletePost(ctx context.Context, postID int64) error {
	zap.L().Info("开始执行DeletePost服务",
		zap.Int64("post_id", postID),
	)

	tracer := otel.Tracer("post-service")
	spanCtx, span := tracer.Start(ctx, "service.DeletePost")
	defer span.End()

	err := p.PostRepo.DeletePost(spanCtx, postID)
	if err != nil {
		zap.L().Error("× 删除文章失败",
			zap.Error(err),
			zap.Int64("post_id", postID),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(attribute.Bool("service.success", true))

	zap.L().Info("√ 删除文章成功",
		zap.Int64("post_id", postID),
	)

	return nil
}

func (p *postServiceImpl) GetPostByID(ctx context.Context, postID int64) (*model.Post, error) {
	zap.L().Info("开始执行GetPostByID服务",
		zap.Int64("post_id", postID),
	)

	tracer := otel.Tracer("post-service")
	spanCtx, span := tracer.Start(ctx, "service.GetPostByID")
	defer span.End()

	post, err := p.PostRepo.GetPostByID(spanCtx, postID)
	if err != nil {
		zap.L().Error("× 获取文章失败",
			zap.Error(err),
			zap.Int64("post_id", postID),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Bool("service.success", true))

	return post, nil
}

func (p *postServiceImpl) GetPostsByOwner(ctx context.Context, ownerID int64) ([]model.Post, error) {
	zap.L().Info("开始执行GetPostsByOwner服务",
		zap.Int64("owner_id", ownerID),
	)

	tracer := otel.Tracer("post-service")
	spanCtx, span := tracer.Start(ctx, "service.GetPostsByOwner")
	defer span.End()

	posts, err := p.PostRepo.GetPostsByOwner(spanCtx, ownerID)
	if err != nil {
		zap.L().Error("× 获取用户文章列表失败",
			zap.Error(err),
			zap.Int64("owner_id", ownerID),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(
		attribute.Int("posts.count", len(posts)),
		attribute.Bool("service.success", true),
	)

	return posts, nil
}
