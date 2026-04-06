package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"microservicesDemo/L8-dtm/internal/user-service/model"
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
	userCacheKeyPrefix = "user:"
	userCacheTTL       = 30 * time.Minute
)

func InitDB(dsn string) (*gorm.DB, error) {
	zap.L().Info("开始初始化数据库连接",
		zap.String("component", "dao"),
	)

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

type UserRepository interface {
	AddUser(ctx context.Context, user *model.User) error
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByID(ctx context.Context, userID int64) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, userID int64) error
}

type userRepositoryImpl struct {
	db    *gorm.DB
	cache *redis.RedisCluster
}

func NewUserRepo(db *gorm.DB, cache *redis.RedisCluster) UserRepository {
	zap.L().Info("开始初始化UserRepository",
		zap.String("component", "dao"),
	)

	err := db.AutoMigrate(&model.User{})
	if err != nil {
		zap.L().Fatal("× 自动迁移User模型失败",
			zap.Error(err),
			zap.String("component", "dao"),
		)
		return nil
	}

	zap.L().Info("√ 自动迁移User模型成功",
		zap.String("component", "dao"),
	)

	zap.L().Info("√ 初始化UserRepository成功",
		zap.String("component", "dao"),
	)
	return &userRepositoryImpl{db: db, cache: cache}
}

func (repo *userRepositoryImpl) AddUser(ctx context.Context, user *model.User) error {
	zap.L().Info("开始执行AddUser操作",
		zap.String("username", user.Username),
		zap.String("email", user.Email),
		zap.String("component", "dao"),
	)

	tracer := otel.Tracer("user-service")
	ctx, span := tracer.Start(ctx, "dao.AddUser")
	defer span.End()

	var existUser model.User
	result := repo.db.Where("username = ?", user.Username).First(&existUser)
	if result.Error == nil {
		zap.L().Error("× 用户名已存在",
			zap.String("component", "dao"),
			zap.String("username", user.Username),
		)
		span.RecordError(errors.New("用户名已存在"))
		span.SetStatus(codes.Error, "用户名已存在")
		span.SetAttributes(attribute.Bool("dao.success", false))
		return errors.New("用户名已存在")
	}

	result = repo.db.Where("email = ?", user.Email).First(&existUser)
	if result.Error == nil {
		zap.L().Error("× 邮箱已存在",
			zap.String("component", "dao"),
			zap.String("email", user.Email),
		)
		span.RecordError(errors.New("邮箱已存在"))
		span.SetStatus(codes.Error, "邮箱已存在")
		span.SetAttributes(attribute.Bool("dao.success", false))
		return errors.New("邮箱已存在")
	}

	userID, err := snowflake.GenerateID()
	if err != nil {
		zap.L().Error("× 生成雪花ID失败",
			zap.Error(err),
			zap.String("component", "dao"),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return errors.New("生成用户ID失败")
	}
	user.UserID = userID

	err = repo.db.Create(user).Error
	if err != nil {
		zap.L().Error("× 创建用户失败",
			zap.Error(err),
			zap.String("component", "dao"),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("dao.success", false))
		return errors.New("创建用户失败: " + err.Error())
	}

	userJSON, _ := json.Marshal(user)
	cacheKey := fmt.Sprintf("%s%d", userCacheKeyPrefix, user.UserID)
	if err := repo.cache.Set(ctx, cacheKey, string(userJSON), userCacheTTL); err != nil {
		zap.L().Warn("× 缓存用户信息失败",
			zap.Error(err),
			zap.Int64("user_id", user.UserID),
		)
	}

	span.SetAttributes(
		attribute.Int64("user.id", user.UserID),
		attribute.Bool("dao.success", true),
	)

	zap.L().Info("√ 创建用户成功",
		zap.Int64("user_id", user.UserID),
		zap.String("username", user.Username),
		zap.String("component", "dao"),
	)

	return nil
}

func (repo *userRepositoryImpl) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	zap.L().Info("开始执行GetUserByUsername操作",
		zap.String("username", username),
		zap.String("component", "dao"),
	)

	tracer := otel.Tracer("user-service")
	ctx, span := tracer.Start(ctx, "dao.GetUserByUsername")
	defer span.End()

	cacheKey := fmt.Sprintf("%susername:%s", userCacheKeyPrefix, username)
	cachedData, err := repo.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var user model.User
		if json.Unmarshal([]byte(cachedData), &user) == nil {
			span.SetAttributes(attribute.Bool("cache.hit", true))
			zap.L().Info("√ 从缓存获取用户成功",
				zap.String("username", username),
				zap.String("component", "dao"),
			)
			return &user, nil
		}
	}

	var user model.User
	result := repo.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			span.SetAttributes(attribute.Bool("dao.found", false))
			return nil, errors.New("用户不存在")
		}
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return nil, result.Error
	}

	userJSON, _ := json.Marshal(user)
	repo.cache.Set(ctx, cacheKey, string(userJSON), userCacheTTL)

	span.SetAttributes(
		attribute.Int64("user.id", user.UserID),
		attribute.Bool("cache.hit", false),
		attribute.Bool("dao.success", true),
	)

	zap.L().Info("√ 从数据库获取用户成功",
		zap.Int64("user_id", user.UserID),
		zap.String("username", username),
		zap.String("component", "dao"),
	)

	return &user, nil
}

func (repo *userRepositoryImpl) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	zap.L().Info("开始执行GetUserByID操作",
		zap.Int64("user_id", userID),
		zap.String("component", "dao"),
	)

	tracer := otel.Tracer("user-service")
	ctx, span := tracer.Start(ctx, "dao.GetUserByID")
	defer span.End()

	cacheKey := fmt.Sprintf("%s%d", userCacheKeyPrefix, userID)
	cachedData, err := repo.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var user model.User
		if json.Unmarshal([]byte(cachedData), &user) == nil {
			span.SetAttributes(attribute.Bool("cache.hit", true))
			zap.L().Info("√ 从缓存获取用户成功",
				zap.Int64("user_id", userID),
				zap.String("component", "dao"),
			)
			return &user, nil
		}
	}

	var user model.User
	result := repo.db.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			span.SetAttributes(attribute.Bool("dao.found", false))
			return nil, errors.New("用户不存在")
		}
		span.RecordError(result.Error)
		span.SetStatus(codes.Error, result.Error.Error())
		return nil, result.Error
	}

	userJSON, _ := json.Marshal(user)
	repo.cache.Set(ctx, cacheKey, string(userJSON), userCacheTTL)

	span.SetAttributes(
		attribute.Bool("cache.hit", false),
		attribute.Bool("dao.success", true),
	)

	zap.L().Info("√ 从数据库获取用户成功",
		zap.Int64("user_id", userID),
		zap.String("component", "dao"),
	)

	return &user, nil
}

func (repo *userRepositoryImpl) UpdateUser(ctx context.Context, user *model.User) error {
	zap.L().Info("开始执行UpdateUser操作",
		zap.Int64("user_id", user.UserID),
		zap.String("component", "dao"),
	)

	tracer := otel.Tracer("user-service")
	ctx, span := tracer.Start(ctx, "dao.UpdateUser")
	defer span.End()

	err := repo.db.Save(user).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		zap.L().Error("× 更新用户失败",
			zap.Error(err),
			zap.Int64("user_id", user.UserID),
		)
		return err
	}

	cacheKey := fmt.Sprintf("%s%d", userCacheKeyPrefix, user.UserID)
	repo.cache.Del(ctx, cacheKey)

	userJSON, _ := json.Marshal(user)
	repo.cache.Set(ctx, cacheKey, string(userJSON), userCacheTTL)

	span.SetAttributes(attribute.Bool("dao.success", true))

	zap.L().Info("√ 更新用户成功",
		zap.Int64("user_id", user.UserID),
		zap.String("component", "dao"),
	)

	return nil
}

func (repo *userRepositoryImpl) DeleteUser(ctx context.Context, userID int64) error {
	zap.L().Info("开始执行DeleteUser操作",
		zap.Int64("user_id", userID),
		zap.String("component", "dao"),
	)

	tracer := otel.Tracer("user-service")
	ctx, span := tracer.Start(ctx, "dao.DeleteUser")
	defer span.End()

	err := repo.db.Delete(&model.User{}, userID).Error
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		zap.L().Error("× 删除用户失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return err
	}

	cacheKey := fmt.Sprintf("%s%d", userCacheKeyPrefix, userID)
	repo.cache.Del(ctx, cacheKey)

	span.SetAttributes(attribute.Bool("dao.success", true))

	zap.L().Info("√ 删除用户成功",
		zap.Int64("user_id", userID),
		zap.String("component", "dao"),
	)

	return nil
}
