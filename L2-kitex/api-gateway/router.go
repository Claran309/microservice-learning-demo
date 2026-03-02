package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"
)

func RegisterRoutes(srv *server.Hertz, handler APIHandler) {
	user := srv.Group("/user")
	user.POST("/register", handler.Register)
}
