package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	KeyName    = "request_id"
	HeaderName = "X-Request-ID"
)

func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, _ := uuid.NewV7()
		ctx.Set(KeyName, id.String())
		ctx.Header(HeaderName, id.String())
		ctx.Next()
	}
}
