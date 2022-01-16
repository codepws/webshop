package middlewares

import (
	"fmt"
	"net/http"
	"webshop-api/user-web/common"
	"webshop-api/user-web/common/errno"
	"webshop-api/user-web/middlewares/jwt"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware 基于JWT的认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URI
		// 这里假设Token放在Header的Authorization中，并使用Bearer开头
		// 这里的具体实现方式要依据你的实际业务情况决定
		authTokenValue := c.Request.Header.Get("X-Token")
		if authTokenValue == "" {
			common.ResponseErrorWithMsg(c, http.StatusUnauthorized, -1, "未登陆", nil)
			c.Abort()
			return
		}

		fmt.Println(authTokenValue)

		// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		mc, err := jwt.UnwrapToken(authTokenValue)
		if err != nil {
			fmt.Println(err)

			if err == errno.ErrTokenExpired {
				common.ResponseErrorWithMsg(c, http.StatusUnauthorized, -1, "授权已过期", nil)
			} else {
				common.ResponseError(c, http.StatusUnauthorized, errno.ErrTokenValidation)
			}

			c.Abort() //用来终止request, 阻止其到达handler 一般情况下用在鉴权与认证的中间件中
			return
		}

		// 将当前请求的 UserClaims 信息保存到请求的上下文c上
		c.Set("KeyCtxContextUserClaims", *mc)

		// 后续的处理函数可以用过c.Get("Token_UserID")来获取当前请求的用户信息
		c.Next()
	}
}
