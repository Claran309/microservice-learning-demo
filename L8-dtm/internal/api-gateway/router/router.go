package router

import (
	"microservicesDemo/L8-dtm/internal/api-gateway/handler"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func InitRouters(h *server.Hertz, httpHandler *handler.HttpHandler) {
	v1 := h.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/register", httpHandler.Register)
			users.POST("/login", httpHandler.Login)
		}

		posts := v1.Group("/posts")
		{
			posts.POST("/create", httpHandler.CreatePost)
			posts.POST("/saga/create", httpHandler.CreatePostWithSaga)
			posts.DELETE("/delete", httpHandler.DeletePost)
		}
	}
}
