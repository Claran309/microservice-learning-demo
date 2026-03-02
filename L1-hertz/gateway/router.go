package gateway

import (
	"microservicesDemo/internal"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func RegisterRoutes(srv *server.Hertz, handler internal.UserHandler, middleware *JWTMiddleware) error {
	user := srv.Group("/user")

	user.POST("/register", handler.Register)
	//其他路由

	return nil
}
