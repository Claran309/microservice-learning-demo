package router

import (
	"microservicesDemo/L4-govern/internal/api-gateway/handler"
	middleware "microservicesDemo/L4-govern/internal/api-gateway/middleware/rateLimit"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func InitRouters(h *server.Hertz, httpHandler *handler.HttpHandler) {
	RateLimitMiddleware := middleware.NewSecurity(1000)

	user := h.Group("/user")
	user.Use(RateLimitMiddleware.UserRateLimitMiddleware())
	user.POST("/register", httpHandler.Register)
	user.POST("/login", httpHandler.Login)

	post := user.Group("/post")
	post.Use(RateLimitMiddleware.UserRateLimitMiddleware())
	post.POST("/create", httpHandler.CreatePost)
	post.DELETE("/delete", httpHandler.DeletePost)
}
