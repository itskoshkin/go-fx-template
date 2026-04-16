package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"go-fx-template/internal/api/cookies"
	"go-fx-template/internal/config"
	"go-fx-template/internal/logger"
	"go-fx-template/internal/services/errors"
)

type WebController struct {
	router      *gin.Engine
	userService UserService
}

func NewWebController(e *gin.Engine, us UserService) *WebController {
	return &WebController{router: e, userService: us}
}

func (ctrl *WebController) RegisterRoutes() {
	ctrl.router.GET("/", ctrl.Index)
	ctrl.router.GET("/login", ctrl.LogIn)
	ctrl.router.GET("/register", ctrl.Register)
	ctrl.router.GET("/verify-email", ctrl.VerifyEmail)
	ctrl.router.POST("/verify-email", ctrl.VerifyEmail)
	ctrl.router.GET("/forgot-password", ctrl.ForgotPassword)
	ctrl.router.GET("/reset-password", ctrl.ResetPassword)
	ctrl.router.GET("/change-password", ctrl.ChangePasswordPage)
	ctrl.router.GET("/delete-account", ctrl.DeleteAccountPage)
	ctrl.router.NoMethod(ctrl.NoMethod)
	ctrl.router.NoRoute(ctrl.NotFound)
}

func (ctrl *WebController) Index(ctx *gin.Context) {
	if !cookies.HasSessionMarker(ctx) {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}

	ctx.HTML(http.StatusOK, "index", nil)
}

func (ctrl *WebController) LogIn(ctx *gin.Context) {
	if cookies.HasSessionMarker(ctx) {
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	initialAuthView := "login"
	if ctx.Query("mode") == "register" {
		initialAuthView = "register"
	}

	initialForgotPassword := false
	switch ctx.Query("forgot") {
	case "1", "true":
		initialForgotPassword = true
	}

	ctx.HTML(http.StatusOK, "sign-in", gin.H{
		"InitialAuthView":       initialAuthView,
		"InitialForgotPassword": initialForgotPassword,
	})
}

func (ctrl *WebController) Register(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, "/login?mode=register")
}

func (ctrl *WebController) VerifyEmail(ctx *gin.Context) {
	switch ctx.Request.Method {
	case http.MethodGet:
		token := strings.TrimSpace(ctx.Query("token"))
		if token == "" {
			ctx.HTML(http.StatusBadRequest, "verify-email", gin.H{
				"Title":     "Invalid verification link",
				"Message":   "This email verification link is missing a token.",
				"HasResult": true,
				"Success":   false,
			})
			return
		}

		ctx.HTML(http.StatusOK, "verify-email", gin.H{
			"Title":     "Confirm email",
			"Message":   "Click the button below to verify your email address.",
			"HasResult": false,
			"Token":     token,
		})

	case http.MethodPost:
		token := strings.TrimSpace(ctx.PostForm("token"))
		if token == "" {
			ctx.HTML(http.StatusBadRequest, "verify-email", gin.H{
				"Title":     "Invalid verification link",
				"Message":   "This email verification link is missing a token.",
				"HasResult": true,
				"Success":   false,
			})
			return
		}

		if err := ctrl.userService.VerifyEmail(ctx, token); err != nil {
			var status int
			var title, message string
			if _, ok := errors.AsType[svcErr.ValidationError](err); ok {
				status = http.StatusBadRequest
				title = "Verification failed"
				message = "This verification link is invalid, expired, or has already been used."
			} else if _, ok = errors.AsType[svcErr.NotFoundError](err); ok {
				status = http.StatusNotFound
				title = "Verification failed"
				message = "This verification link no longer matches an account."
			} else {
				logger.ErrorWithID(ctx, "failed to verify email in web flow: %v", err)
				status = http.StatusInternalServerError
				title = "Verification unavailable"
				message = fmt.Sprintf("We could not verify your email right now. Your request ID is %s.", ctx.GetString("request_id"))
			}
			ctx.HTML(status, "verify-email", gin.H{
				"Title":     title,
				"Message":   message,
				"HasResult": true,
				"Success":   false,
				"Token":     token,
			})
			return
		}

		ctx.HTML(http.StatusOK, "verify-email", gin.H{
			"Title":     "Email confirmed",
			"Message":   "Your email address has been successfully verified.",
			"HasResult": true,
			"Success":   true,
			"Token":     token,
		})
	}
}

func (ctrl *WebController) ForgotPassword(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, "/login?forgot=1")
}

func (ctrl *WebController) ResetPassword(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.HTML(http.StatusBadRequest, "reset-password", gin.H{
			"Title":    "Invalid reset link",
			"Message":  "This password reset link is missing a token.",
			"HasToken": false,
		})
		return
	}

	ctx.HTML(http.StatusOK, "reset-password", gin.H{
		"Title":    "Set a new password",
		"Message":  "Choose a new password for your account.",
		"HasToken": true,
		"Token":    token,
	})
}

func (ctrl *WebController) ChangePasswordPage(ctx *gin.Context) {
	if !cookies.HasSessionMarker(ctx) {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}

	ctx.HTML(http.StatusOK, "change-password", nil)
}

func (ctrl *WebController) DeleteAccountPage(ctx *gin.Context) {
	if !cookies.HasSessionMarker(ctx) {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}

	ctx.HTML(http.StatusOK, "delete-account", nil)
}

func (ctrl *WebController) NotFound(ctx *gin.Context) {
	if strings.HasPrefix(ctx.Request.URL.Path, viper.GetString(config.ApiBasePath)) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "route not found",
		})
		return
	}

	ctx.HTML(http.StatusNotFound, "error", gin.H{
		"ErrorTitle":       "404 Not Found",
		"ErrorDescription": fmt.Sprintf("Route %s leads to no page.", ctx.Request.URL.EscapedPath()),
	})
}

func (ctrl *WebController) NoMethod(ctx *gin.Context) {
	if strings.HasPrefix(ctx.Request.URL.Path, viper.GetString(config.ApiBasePath)) {
		ctx.JSON(http.StatusMethodNotAllowed, gin.H{
			"message": "route has no method " + ctx.Request.Method,
		})
		return
	}

	ctx.HTML(http.StatusMethodNotAllowed, "error", gin.H{
		"ErrorTitle":       "405 Method Not Allowed",
		"ErrorDescription": fmt.Sprintf("Method %s not allowed on %s.", ctx.Request.Method, ctx.Request.URL.EscapedPath()),
	})
}
