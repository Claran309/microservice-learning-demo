package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"microservicesDemo/L8-dtm/internal/post-service/model"
	"microservicesDemo/L8-dtm/pkg/cache/redis"
	"microservicesDemo/L8-dtm/pkg/id/snowflake"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	postCacheKeyPrefix = "post:"
	postCacheTTL       = 30 * time.Minute
)

func InitDB(dsn string) (*gorm.DB, error) {
	zap.L().Info("开始初始化数据库连接")

	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		zap.L().Error("× 数据库连接失败",
			zap.Error(err),
			zap.String("component", "dao"),
		)
		return nil, errors.New("数据库连接失败：" + err.Error())
	}

	zap.L().Info("√ 数据库连接成功",
		zap.String("component", "dao"),
	)
	return db, nil
}

type PostRepository interface {
	AddPost(ctx context.Context, post *model.Post) error
	GetPostByID(ctx context.Context, postID int64) (*model.Post, error)
	DeletePost(ctx context.Context, postID int64) error
	GetPostsByOwner(ctx context.Context, ownerID int64) ([]model.Post, error)
}

type postRepositoryImpl struct {
	db    *gorm.DB
	cache *redis.RedisCluster
}

func NewPostRepositoryImpl(db *gorm.DB, cache *redis.RedisCluster) PostRepository {
	zap.L().Info("开始初始化PostRepository")

	err := db.AutoMigrate(&model.Post{})
	if err != nil {
		zap.L().Fatal("× 自动迁移Post模型失败",
			zap.Error(err),
			zap.String("component", "dao"),
		)
		return nil
	}

	zap.L().Info("√ 自动迁移Post模型成功",
		zap.String("component", "dao"),
	)

	zap.L().Info("√ 初始化PostRepository成功",
		zap.String("component", "dao"),
	)
	return &postRepositoryImpl{db: db, cache: cache}
}

func (p *postRepositoryImpl) AddPost(ctx context.Context, post *model.Post) error {
	zap.L().Info("开始执行AddPost操作",
		zap.Int64("user_id", post.Owner),
		zap.String("title", post.Title),
		zap.String("component", "dao"),
	)

	tracer := otel.Tracer("post-service")
	ctx, span := tracer.Start(ctx, "dao.AddPost")
	defer span.End()

	postID, err := snowflake.GenerateID()
	if err != nil {
		zap.L().Error("× 生成雪花ID失败",
			zap.Error(err),
			zap.String("component", "dao"),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return errors.New("生成文章ID失败")
	}
	post.PostID = postID

	err = p.db.Create(post).Error
	if err != nil {
		zap.L().Error("× 添加文章失败",
			zap.Error(err),
			zap.String("component", "dao"),
			zap.Int64("user_id", post.Owner),
			zap.String("title", post.Title),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	postJSON, _ := json.Marshal(post)
	cacheKey := fmt.Sprintf("%s%d", postCacheKeyPrefix, post.PostID)
	p.cache.Set(ctx, cacheKey, string(postJSON), postCacheTTL)

	span.SetAttributes(
		attribute.Int64("post.id", post.PostID),
		attribute.Bool("dao.success", true),
	)

	zap.L().Info("√ 添加文章成功",
		zap.Int64("post_id", post.PostID),
		zap.Int64("user_id", post.Owner),
		zap.String("title", post.Title),
		zap.String("component", "dao"),
	)

	return nil
}

func (p *postRepositoryImpl) GetPostByID(ctx context.Context, postID int64) (*model.Post, error) {
	zap.L().Info("开始执行GetPostByID操作",
		zap.Int64("post_id", postID),
		zap.String("component", "dao"),
	)

	tracer := otel.Tracer("post-service")
	ctx, span := tracer.Start(ctx, "dao.GetPostByID")
	defer span.End()

	cacheKey := fmt.Sprintf("%s%d", postCacheKeyPrefix, postID)
	cachedData, err := p.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var post model.Post
		if json.Unmarshal([]byte(cachedData), &post) == nil {
			span.SetAttributes(attribute.Bool("cache.hit", true))
			zap.L().Info("√ 从缓存获取文章成功",
				zap.Int64("post_id", postID),
				zap.String("component", "dao"),
			)
			return &post, nil
		}
	}

	var post model.Post
	result := p.db.First(&post, postID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			span.SetAttributes(attribute.Bool("dao.found", false))
			return nil, errors.New("文章不存在")
		}
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return nil, result.Error
	}

	postJSON, _ := json.Marshal(post)
	p.cache.Set(ctx, cacheKey, string(postJSON), postCacheTTL)

	span.SetAttributes(
		attribute.Bool("cache.hit", false),
		attribute.Bool("dao.success", true),
	)

	zap.L().Info("√ 从数据库获取文章成功",
		zap.Int64("post_id", postID),
		zap.String("component", "dao"),
	)

	return &post, nil
}

func (p *postRepositoryImpl) DeletePost(ctx context.Context, postID int64) error {
	zap.L().Info("开始执行DeletePost操作",
		zap.Int64("post_id", postID),
		zap.String("component", "dao"),
	)

	tracer := otel.Tracer("post-service")
	ctx, span := tracer.Start(ctx, "dao.DeletePost")
	defer span.End()

	err := p.db.Delete(&model.Post{}, postID).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		zap.L().Error("× 删除文章失败",
			zap.Error(err),
			zap.Int64("post_id", postID),
		)
		return err
	}

	cacheKey := fmt.Sprintf("%s%d", postCacheKeyPrefix, postID)
	p.cache.Del(ctx, cacheKey)

	span.SetAttributes(attribute.Bool("dao.success", true))

	zap.L().Info("√ 删除文章成功",
		zap.Int64("post_id", postID),
		zap.String("component", "dao"),
	)

	return nil
}

func (p *postRepositoryImpl) GetPostsByOwner(ctx context.Context, ownerID int64) ([]model.Post, error) {
	zap.L().Info("开始执行GetPostsByOwner操作",
		zap.Int64("owner_id", ownerID),
		zap.String("component", "dao"),
	)

	tracer := otel.Tracer("post-service")
	ctx, span := tracer.Start(ctx, "dao.GetPostsByOwner")
	defer span.End()

	var posts []model.Post
	result := p.db.Where("owner = ?", ownerID).Find(&posts)
	if result.Error != nil {
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return nil, result.Error
	}

	span.SetAttributes(
		attribute.Int("posts.count", len(posts)),
		attribute.Bool("dao.success", true),
	)

	zap.L().Info("√ 获取用户文章列表成功",
		zap.Int64("owner_id", ownerID),
		zap.Int("count", len(posts)),
		zap.String("component", "dao"),
	)

	return posts, nil
}
