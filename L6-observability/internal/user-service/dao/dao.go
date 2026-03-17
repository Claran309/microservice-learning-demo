package dao

import (
	"context"
	"errors"
	"log"
	"microservicesDemo/L6-observability/internal/user-service/model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

type UserRepository interface {
	AddUser(ctx context.Context, user *model.User) error
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
}

type userRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepository {
	err := db.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatal("自动迁移User模型失败")
		return nil
	}

	return &userRepositoryImpl{db: db}
}

func (repo *userRepositoryImpl) AddUser(ctx context.Context, user *model.User) error {
	tracer := otel.Tracer("user-service")
	ctx, span := tracer.Start(ctx, "dao.AddUser")
	defer span.End()
	//sCtx := trace.SpanContextFromContext(ctx)
	//log.Printf("dao entered, TraceID: %v", sCtx.TraceID())

	var exist bool
	repo.db.Where("username = ?", user.Username).First(&exist)
	if exist != false {
		span.RecordError(errors.New("用户名已存在"))
		span.SetStatus(codes.Error, "用户名已存在")
		span.SetAttributes(attribute.Bool("dao.success", false))
		return errors.New("用户名已存在")
	}

	exist = false
	repo.db.Where("email = ?", user.Email).First(&exist)
	if exist != false {
		span.RecordError(errors.New("邮箱已存在"))
		span.SetStatus(codes.Error, "邮箱已存在")
		span.SetAttributes(attribute.Bool("dao.success", false))
		return errors.New("邮箱已存在")
	}

	span.SetAttributes(attribute.Bool("dao.success", true))

	return repo.db.Create(user).Error
}

func (repo *userRepositoryImpl) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	tracer := otel.Tracer("user-service")
	ctx, span := tracer.Start(ctx, "dao.GetUserByUsername")
	defer span.End()

	var user model.User
	repo.db.Where("username = ?", username).First(&user)

	span.SetAttributes(attribute.Bool("dao.success", true))

	return &user, nil
}
