package controllers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"go-fx-template/internal/api/cookies"
	"go-fx-template/internal/api/middlewares"
	"go-fx-template/internal/api/models"
	"go-fx-template/internal/config"
	"go-fx-template/internal/models"
	"go-fx-template/internal/services/errors"
	"go-fx-template/internal/utils/noop"
)

type AuthService interface {
	GenerateTokens(ctx context.Context, userID uuid.UUID) (string, string, error)
	RotateRefreshToken(ctx context.Context, refreshToken string, userID uuid.UUID) (string, string, error)
	ValidateAccessToken(ctx context.Context, token string) (uuid.UUID, error)
	ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, error)
	RevokeAuthTokens(ctx context.Context, accessToken, refreshToken string) error
}

type UserService interface {
	Register(ctx context.Context, req apiModels.RegisterUserRequest) (models.User, error)
	VerifyEmail(ctx context.Context, token string) error
	LogIn(ctx context.Context, req apiModels.LogInUserRequest) (models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error)
	UpdateUserByID(ctx context.Context, id uuid.UUID, req apiModels.UpdateUserRequest) error
	UpdateAvatar(ctx context.Context, id uuid.UUID, avatarURL string) error
	DeleteAvatar(ctx context.Context, id uuid.UUID) error
	VerifyPassword(ctx context.Context, id uuid.UUID, password string) error
	ChangePassword(ctx context.Context, id uuid.UUID, req apiModels.ChangePasswordRequest) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserController struct {
	router       *gin.Engine
	mw           *middlewares.Middlewares
	authService  AuthService
	userService  UserService
	imageService ImageService
}

func NewUserController(e *gin.Engine, mw *middlewares.Middlewares, as AuthService, us UserService, imgSvc ImageService) *UserController {
	return &UserController{router: e, mw: mw, authService: as, userService: us, imageService: imgSvc}
}

func (ctrl *UserController) RegisterRoutes() {
	basePath := ctrl.router.Group(viper.GetString(config.ApiBasePath))
	authRoutes := basePath.Group("/auth")
	{
		authRoutes.POST("/register", ctrl.Register)
		authRoutes.POST("/verify-email", ctrl.VerifyEmail)
		authRoutes.POST("/login", ctrl.LogIn)
		authRoutes.POST("/refresh", ctrl.RefreshTokens)
		authRoutes.POST("/logout", ctrl.mw.AuthMiddleware(), ctrl.LogOut)

		authRoutes.POST("/forgot-password", ctrl.ForgotPassword)
		authRoutes.POST("/set-new-password", ctrl.SetNewPassword)
	}
	userRoutes := basePath.Group("/users")
	{
		authedUserRoutes := userRoutes.Group("").Use(ctrl.mw.AuthMiddleware())
		{
			authedUserRoutes.GET("/me", ctrl.GetCurrentUser)
			authedUserRoutes.PATCH("/me", ctrl.UpdateCurrentUser)
			authedUserRoutes.PUT("/me/avatar", ctrl.UpdateAvatar)
			authedUserRoutes.DELETE("/me/avatar", ctrl.DeleteAvatar)
			authedUserRoutes.PATCH("/me/update-password", ctrl.UpdateCurrentPassword)
			authedUserRoutes.DELETE("/me", ctrl.DeleteCurrentUser)
		}
	}
}

// Register
// @Summary Register a new user
// @Description Creates a user account and returns auth tokens with the created user.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body apiModels.RegisterUserRequest true "Registration payload"
// @Success 201 {object} apiModels.AuthResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 409 {object} apiModels.ConflictError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/register [post]
func (ctrl *UserController) Register(ctx *gin.Context) {
	var req apiModels.RegisterUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	user, err := ctrl.userService.Register(ctx, req)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	accessToken, refreshToken, err := ctrl.authService.GenerateTokens(ctx, user.ID)
	if err != nil {
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	cookies.SetRefreshTokenCookie(ctx, refreshToken, viper.GetDuration(config.RefreshTokenTTL))
	cookies.SetSessionMarkerCookie(ctx, viper.GetDuration(config.RefreshTokenTTL))

	ctx.JSON(http.StatusCreated, apiModels.AuthResponse{
		AuthTokensResponse: apiModels.AuthTokensResponse{AccessToken: accessToken},
		User:               apiModels.ToUserResponse(user),
	})
}

// VerifyEmail
// @Summary Verify email address
// @Description Verifies a user's email using a verification token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body apiModels.VerifyEmailRequest true "Verification payload"
// @Success 200 {object} apiModels.APIResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/verify-email [post]
func (ctrl *UserController) VerifyEmail(ctx *gin.Context) {
	var req apiModels.VerifyEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	if err := ctrl.userService.VerifyEmail(ctx, req.Token); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, apiModels.APIResponse{Message: "email verified"})
}

// LogIn
// @Summary Log in a user
// @Description Authenticates a user and returns a fresh token pair with user data.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body apiModels.LogInUserRequest true "Login payload"
// @Success 200 {object} apiModels.AuthResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 403 {object} apiModels.ForbiddenError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/login [post]
func (ctrl *UserController) LogIn(ctx *gin.Context) {
	var req apiModels.LogInUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	user, err := ctrl.userService.LogIn(ctx, req)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	accessToken, refreshToken, err := ctrl.authService.GenerateTokens(ctx, user.ID)
	if err != nil {
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	cookies.SetRefreshTokenCookie(ctx, refreshToken, viper.GetDuration(config.RefreshTokenTTL))
	cookies.SetSessionMarkerCookie(ctx, viper.GetDuration(config.RefreshTokenTTL))

	ctx.JSON(http.StatusOK, apiModels.AuthResponse{
		AuthTokensResponse: apiModels.AuthTokensResponse{AccessToken: accessToken},
		User:               apiModels.ToUserResponse(user),
	})
}

// RefreshTokens
// @Summary Refresh auth tokens
// @Description Issues a new access and refresh token pair from a valid refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body apiModels.RefreshTokenRequest false "Optional refresh token payload for non-browser clients; browsers use the HttpOnly cookie"
// @Success 200 {object} apiModels.AuthTokensResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/refresh [post]
func (ctrl *UserController) RefreshTokens(ctx *gin.Context) {
	refreshToken, ok, err := readRefreshTokenFromRequest(ctx)
	if err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	if !ok {
		apiModels.Error(ctx, http.StatusUnauthorized, "refresh token is invalid or expired")
		return
	}

	userID, err := ctrl.authService.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, svcErr.ErrRefreshTokenCompromised) {
			cookies.ClearAuthCookies(ctx)
			apiModels.Error(ctx, http.StatusUnauthorized, "refresh token is compromised")
			return
		}
		cookies.ClearAuthCookies(ctx)
		apiModels.Error(ctx, http.StatusUnauthorized, "refresh token is invalid or expired")
		return
	}

	accessToken, nextRefreshToken, err := ctrl.authService.RotateRefreshToken(ctx, refreshToken, userID)
	if err != nil {
		if errors.Is(err, svcErr.ErrRefreshTokenCompromised) {
			cookies.ClearAuthCookies(ctx)
			apiModels.Error(ctx, http.StatusUnauthorized, "refresh token is compromised")
			return
		}
		if errors.Is(err, svcErr.ErrRefreshTokenInvalidOrExpired) {
			cookies.ClearAuthCookies(ctx)
			apiModels.Error(ctx, http.StatusUnauthorized, "refresh token is invalid or expired")
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	cookies.SetRefreshTokenCookie(ctx, nextRefreshToken, viper.GetDuration(config.RefreshTokenTTL))
	cookies.SetSessionMarkerCookie(ctx, viper.GetDuration(config.RefreshTokenTTL))

	ctx.JSON(http.StatusOK, apiModels.AuthTokensResponse{
		AccessToken: accessToken,
	})
}

// LogOut
// @Summary Log out a user
// @Description Revokes the current access token and the provided refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param request body apiModels.RefreshTokenRequest false "Optional refresh token payload for non-browser clients; browsers use the HttpOnly cookie"
// @Success 200 {object} apiModels.APIResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/logout [post]
func (ctrl *UserController) LogOut(ctx *gin.Context) {
	refreshToken, _, err := readRefreshTokenFromRequest(ctx)
	if err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	accessToken, found := strings.CutPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	if !found {
		apiModels.Error(ctx, http.StatusUnauthorized, "invalid or expired access token")
		return
	}

	if err = ctrl.authService.RevokeAuthTokens(ctx, accessToken, refreshToken); err != nil {
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	cookies.ClearAuthCookies(ctx)

	ctx.JSON(http.StatusOK, apiModels.APIResponse{Message: "logged out"})
}

func readRefreshTokenFromRequest(ctx *gin.Context) (string, bool, error) {
	var req apiModels.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		return "", false, err
	}

	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		refreshToken = cookies.ReadRefreshTokenCookie(ctx)
	}

	return refreshToken, refreshToken != "", nil
}

// ForgotPassword
// @Summary Request password reset
// @Description Sends a password reset link if an account with the email exists.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body apiModels.ForgotPasswordRequest true "Forgot password payload"
// @Success 200 {object} apiModels.APIResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/forgot-password [post]
func (ctrl *UserController) ForgotPassword(ctx *gin.Context) {
	var req apiModels.ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	_ = ctrl.userService.RequestPasswordReset(ctx, req.Email) // Always return same response to prevent email enumeration

	ctx.JSON(http.StatusOK, apiModels.APIResponse{Message: "if account with this email exists, you will receive a password reset link shortly"})
}

// SetNewPassword
// @Summary Set a new password
// @Description Resets a user's password using a valid reset token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body apiModels.SetNewPasswordRequest true "Reset password payload"
// @Success 200 {object} apiModels.APIResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 500 {object} apiModels.APIError
// @Router /auth/set-new-password [post]
func (ctrl *UserController) SetNewPassword(ctx *gin.Context) {
	var req apiModels.SetNewPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	if err := ctrl.userService.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, apiModels.APIResponse{Message: "password has been reset"})
}

// GetCurrentUser
// @Summary Get current user
// @Description Returns the currently authenticated user's profile.
// @Tags Users
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Success 200 {object} apiModels.UserResponse
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 500 {object} apiModels.APIError
// @Router /users/me [get]
func (ctrl *UserController) GetCurrentUser(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	user, err := ctrl.userService.GetUserByID(ctx, userID)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, apiModels.ToUserResponse(user))
}

// UpdateCurrentUser
// @Summary Update current user
// @Description Updates the currently authenticated user's profile fields.
// @Tags Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param request body apiModels.UpdateUserRequest true "Profile update payload"
// @Success 200 {object} apiModels.UserResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 409 {object} apiModels.ConflictError
// @Failure 500 {object} apiModels.APIError
// @Router /users/me [patch]
func (ctrl *UserController) UpdateCurrentUser(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	var req apiModels.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	if err := ctrl.userService.UpdateUserByID(ctx, userID, req); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	user, err := ctrl.userService.GetUserByID(ctx, userID)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, apiModels.ToUserResponse(user))
}

// UpdateAvatar
// @Summary Update current user avatar
// @Description Uploads and sets a new avatar for the authenticated user.
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param avatar formData file true "Avatar image"
// @Success 200 {object} apiModels.UserResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 500 {object} apiModels.APIError
// @Router /users/me/avatar [put]
func (ctrl *UserController) UpdateAvatar(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	fileHeader, err := ctx.FormFile("avatar")
	if err != nil {
		apiModels.Error(ctx, http.StatusBadRequest, "avatar file is required")
		return
	}

	// Note: if behind nginx, ensure client_max_body_size matches MinioMaxFileSize
	if fileHeader.Size > viper.GetInt64(config.MinioMaxFileSize)<<20 { // 10 << 20 = 10 * 1 048 576 = 10 485 760 bytes = 10 MB
		apiModels.Error(ctx, http.StatusBadRequest, fmt.Sprintf("image file too large (max %d MB)", viper.GetInt(config.MinioMaxFileSize)))
		return
	}

	switch fileHeader.Header.Get("Content-Type") {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		noop.Continue()
	default:
		apiModels.Error(ctx, http.StatusBadRequest, "unsupported image format (PNG, JPG, WEBP or GIF only)")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}
	defer func() { _ = file.Close() }()

	avatarURL, err := ctrl.imageService.Upload(ctx, "user", userID, file, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	if err = ctrl.userService.UpdateAvatar(ctx, userID, avatarURL); err != nil {
		_ = ctrl.imageService.Delete(ctx, avatarURL)
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	user, err := ctrl.userService.GetUserByID(ctx, userID)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, apiModels.ToUserResponse(user))
}

// DeleteAvatar
// @Summary Delete current user avatar
// @Description Removes the current avatar from the authenticated user's profile.
// @Tags Users
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Success 204 "No Content"
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 500 {object} apiModels.APIError
// @Router /users/me/avatar [delete]
func (ctrl *UserController) DeleteAvatar(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	user, err := ctrl.userService.GetUserByID(ctx, userID)
	if err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	if user.Avatar == nil {
		ctx.Status(http.StatusNoContent)
		return
	}

	if err = ctrl.userService.DeleteAvatar(ctx, userID); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	_ = ctrl.imageService.Delete(ctx, *user.Avatar)

	ctx.Status(http.StatusNoContent)
}

// UpdateCurrentPassword
// @Summary Change current user password
// @Description Changes the authenticated user's password after verifying the current password.
// @Tags Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param request body apiModels.ChangePasswordRequest true "Change password payload"
// @Success 200 {object} apiModels.APIResponse
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 403 {object} apiModels.ForbiddenError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 500 {object} apiModels.APIError
// @Router /users/me/update-password [patch]
func (ctrl *UserController) UpdateCurrentPassword(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	var req apiModels.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	if err := ctrl.userService.ChangePassword(ctx, userID, req); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, apiModels.APIResponse{Message: "password changed"})
}

// DeleteCurrentUser
// @Summary Delete current user account
// @Description Deletes the authenticated user's account after password verification.
// @Tags Users
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param request body apiModels.DeleteAccountRequest true "Delete account payload"
// @Success 204 "No Content"
// @Failure 400 {object} apiModels.BadRequestError
// @Failure 401 {object} apiModels.UnauthorizedError
// @Failure 403 {object} apiModels.ForbiddenError
// @Failure 404 {object} apiModels.NotFoundError
// @Failure 500 {object} apiModels.APIError
// @Router /users/me [delete]
func (ctrl *UserController) DeleteCurrentUser(ctx *gin.Context) {
	userID, ok := middlewares.GetUserID(ctx)
	if !ok {
		apiModels.RespondWithInternalError(ctx, "user id not found in context")
		return
	}

	var req apiModels.DeleteAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apiModels.RespondWithBindError(ctx, err)
		return
	}

	if err := ctrl.userService.VerifyPassword(ctx, userID, req.CurrentPassword); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	if err := ctrl.userService.Delete(ctx, userID); err != nil {
		if apiModels.RespondWithServiceError(ctx, err) {
			return
		}
		apiModels.RespondWithInternalError(ctx, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}
