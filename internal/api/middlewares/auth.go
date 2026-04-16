package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-fx-template/internal/api/models"
)

const (
	UserIDKey = "user_id"
)

type AuthService interface {
	ValidateAccessToken(ctx context.Context, token string) (uuid.UUID, error)
}

type Middlewares struct{ authService AuthService }

func newMiddlewares(as AuthService) *Middlewares { return &Middlewares{authService: as} }

func NewMiddlewares(as AuthService) *Middlewares { return newMiddlewares(as) }

func (mw *Middlewares) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" {
			apiModels.Error(ctx, http.StatusUnauthorized, "missing authorization header")
			return
		}
		token, found := strings.CutPrefix(header, "Bearer ")
		if !found {
			apiModels.Error(ctx, http.StatusUnauthorized, "invalid authorization format")
			return
		}
		userID, err := mw.authService.ValidateAccessToken(ctx, token)
		if err != nil {
			apiModels.Error(ctx, http.StatusUnauthorized, "access token is invalid or expired")
			return
		}
		ctx.Set(UserIDKey, userID)
		ctx.Next()
	}
}

func (mw *Middlewares) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" {
			ctx.Next()
			return
		}
		token, found := strings.CutPrefix(header, "Bearer ")
		if !found {
			ctx.Next()
			return
		}
		userID, err := mw.authService.ValidateAccessToken(ctx, token)
		if err != nil {
			ctx.Next()
			return
		}
		ctx.Set(UserIDKey, userID)
		ctx.Next()
	}
}

func GetUserID(ctx *gin.Context) (uuid.UUID, bool) {
	value, exists := ctx.Get(UserIDKey)
	if !exists {
		return uuid.UUID{}, false
	}
	userID, ok := value.(uuid.UUID)
	return userID, ok
}
