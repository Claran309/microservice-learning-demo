package internal

import (
	"github.com/cloudwego/hertz/pkg/app/server"
)

func InitRouters(h *server.Hertz, httpHandler *HttpHandler) {
	user := h.Group("/user")
	user.POST("/register", httpHandler.Register)
	user.POST("/login", httpHandler.Login)

	post := user.Group("/post")
	post.POST("/create", httpHandler.CreatePost)
	post.DELETE("/delete", httpHandler.DeletePost)
}
