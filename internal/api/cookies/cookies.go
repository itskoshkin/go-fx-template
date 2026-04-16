package cookies

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	AccessTokenCookieName  = "go_fx_template.access_token"
	RefreshTokenCookieName = "go_fx_template.refresh_token"
	AuthSessionCookieName  = "go_fx_template.session"
)

func SetRefreshTokenCookie(ctx *gin.Context, refreshToken string, ttl time.Duration) {
	setCookie(ctx, RefreshTokenCookieName, refreshToken, ttl, true)
}

func SetSessionMarkerCookie(ctx *gin.Context, ttl time.Duration) {
	setCookie(ctx, AuthSessionCookieName, "1", ttl, false)
}

func HasSessionMarker(ctx *gin.Context) bool {
	accessToken, _ := ctx.Cookie(AccessTokenCookieName)
	sessionMarker, _ := ctx.Cookie(AuthSessionCookieName)
	return accessToken != "" || sessionMarker != ""
}

func ReadRefreshTokenCookie(ctx *gin.Context) string {
	refreshToken, _ := ctx.Cookie(RefreshTokenCookieName)
	return strings.TrimSpace(refreshToken)
}

func ClearAuthCookies(ctx *gin.Context) {
	clearCookie(ctx, AccessTokenCookieName, false)
	clearCookie(ctx, RefreshTokenCookieName, true)
	clearCookie(ctx, AuthSessionCookieName, false)
}

func setCookie(ctx *gin.Context, name, value string, ttl time.Duration, httpOnly bool) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(name, value, int(ttl.Seconds()), "/", "", isSecureRequest(ctx), httpOnly)
}

func clearCookie(ctx *gin.Context, name string, httpOnly bool) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(name, "", -1, "/", "", isSecureRequest(ctx), httpOnly)
}

func isSecureRequest(ctx *gin.Context) bool {
	if ctx.Request.TLS != nil {
		return true
	}

	return strings.EqualFold(strings.TrimSpace(ctx.GetHeader("X-Forwarded-Proto")), "https")
}
