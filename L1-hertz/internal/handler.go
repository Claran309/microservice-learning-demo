package internal

import (
	"context"
	util "microservicesDemo/utils"

	"github.com/cloudwego/hertz/pkg/app"
)

type UserHandler struct {
	UserService *UserService
}

func NewUserHandler(userService *UserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Register(c context.Context, ctx *app.RequestContext) {
	var req RegisterRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		util.Error(ctx, 400, err.Error())
		return
	}

	user, err := h.UserService.Register(req.Username, req.Password)

	if err != nil {
		util.Error(ctx, 500, err.Error())
		return
	}

	//返回响应
	util.Success(ctx, map[string]any{
		"username": user.Name,
	}, "RegisterRequest registered successfully")
}
