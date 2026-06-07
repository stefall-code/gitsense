package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminAuth 验证 /admin/* 路由的 Bearer Token
// 从环境变量 ADMIN_TOKEN 读取，请求头需带 Authorization: Bearer <token>
func AdminAuth(adminToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 未配置 ADMIN_TOKEN 时拒绝所有请求
		if adminToken == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": gin.H{
					"code":    "ADMIN_NOT_CONFIGURED",
					"message": "ADMIN_TOKEN not configured on server",
				},
			})
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header required",
				},
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] != adminToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "invalid or missing admin token",
				},
			})
			return
		}

		c.Next()
	}
}
