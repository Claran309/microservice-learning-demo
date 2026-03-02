package main

import (
	"microservicesDemo/gateway"
	"microservicesDemo/internal"
	"microservicesDemo/utils/jwt_util"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	h := server.Default()

	var config = jwt_util.Config{
		Issuer:         "demo",
		SecretKey:      "demo",
		ExpirationTime: time.Hour * 24 * 7,
	}

	userRepo := internal.UserRepo{}
	userService := internal.UserService{userRepo}
	userHandler := internal.UserHandler{&userService}

	jwtUtil := jwt_util.NewJWTUtil(&config)
	JWTMiddleware := gateway.NewJWTMiddleware(jwtUtil)

	err := gateway.RegisterRoutes(h, userHandler, JWTMiddleware)
	if err != nil {
		panic(err)
	}

	h.Spin()
}
