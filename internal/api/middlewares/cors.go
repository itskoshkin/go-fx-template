package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOriginsList ...string) gin.HandlerFunc {
	allowedOrigins := make(map[string]struct{}, len(allowedOriginsList))
	for _, origin := range allowedOriginsList {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowedOrigins[origin] = struct{}{}
		}
	}

	return func(ctx *gin.Context) {
		requestOrigin := strings.TrimSpace(ctx.GetHeader("Origin"))
		if len(allowedOrigins) == 0 {
			ctx.Header("Access-Control-Allow-Origin", "*")
		} else if requestOrigin != "" {
			if _, ok := allowedOrigins[requestOrigin]; ok {
				ctx.Header("Access-Control-Allow-Origin", requestOrigin)
				ctx.Writer.Header().Add("Vary", "Origin")
				// ctx.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Header("Access-Control-Max-Age", "86400") // 24h

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
