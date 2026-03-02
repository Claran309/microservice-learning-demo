package gateway

import (
	"context"
	"errors"
	util "microservicesDemo/utils"
	"microservicesDemo/utils/jwt_util"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/golang-jwt/jwt/v5"
)

type JWTMiddleware struct {
	jwtUtil jwt_util.Util
}

func NewJWTMiddleware(jwtUtil jwt_util.Util) *JWTMiddleware {
	return &JWTMiddleware{
		jwtUtil: jwtUtil,
	}
}

// JWTAuthentication 进行jwt认证
func (m *JWTMiddleware) JWTAuthentication() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		authorizationHeader := string(ctx.GetHeader("Authorization"))
		if authorizationHeader == "" {
			util.Error(ctx, 401, "未登录！") // 未登录
			ctx.Abort()
			return
		}

		parts := strings.SplitN(authorizationHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			util.Error(ctx, 401, "未登录！")
			ctx.Abort()
			return
		}

		tokenString := parts[1]

		token, err := m.jwtUtil.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				util.Error(ctx, 401, "Token is expired")
				ctx.Abort()
				return
			}
			util.Error(ctx, 401, "Token is invalid:"+err.Error())
			ctx.Abort()
			return
		}

		claims, err := m.jwtUtil.ExtractClaims(token)
		if err != nil {
			util.Error(ctx, 500, "Failed to extract claims")
			ctx.Abort()
			return
		}
		//fmt.Println("========================================")
		if userIDFloat, ok := claims["user_id"].(float64); ok {
			// 安全转换：float64 转 int
			userID := int(userIDFloat)
			ctx.Set("user_id", userID)
		} else if userIDInt, ok := claims["user_id"].(int); ok {
			// 如果已经是 int
			ctx.Set("user_id", userIDInt)
		} else {
			util.Error(ctx, 401, "无效的 user_id 类型")
			ctx.Abort()
			return
		}
		ctx.Set("username", claims["username"])
		ctx.Set("role", claims["role"])
		ctx.Next(c)
	}
}
