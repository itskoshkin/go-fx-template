package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"go-fx-template/internal/api/controllers"
	"go-fx-template/internal/api/models"
	"go-fx-template/internal/config"
	"go-fx-template/internal/logger"
	"go-fx-template/internal/models"
	"go-fx-template/internal/repository/cache/errors"
	"go-fx-template/internal/services/errors"
	"go-fx-template/internal/utils/crypto"
)

type UserStorage interface {
	CreateUser(ctx context.Context, user models.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error)
	GetUserByUsername(ctx context.Context, username string) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	UpdateUserByID(ctx context.Context, id uuid.UUID, user models.User, fields []string) error
	UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error
	RemoveAvatar(ctx context.Context, id uuid.UUID) error
	SetUserEmailAsVerified(ctx context.Context, id uuid.UUID) error
	DeleteUserByID(ctx context.Context, id uuid.UUID) error
}

type EmailService interface {
	SendPasswordResetLetter(ctx context.Context, to, token string) error
	SendEmailVerificationLetter(ctx context.Context, to, token string) error
}

type UserServiceImpl struct {
	tokenStorage TokenStorage
	userStorage  UserStorage
	emailService EmailService
}

func newUserService(us UserStorage, ts TokenStorage, es EmailService) *UserServiceImpl {
	return &UserServiceImpl{userStorage: us, tokenStorage: ts, emailService: es}
}

func NewUserService(us UserStorage, ts TokenStorage, es EmailService) controllers.UserService {
	return newUserService(us, ts, es)
}

func (svc *UserServiceImpl) Register(ctx context.Context, req apiModels.RegisterUserRequest) (models.User, error) {
	if err := req.Validate(); err != nil {
		return models.User{}, svcErr.ValidationError{Message: "validation failed: " + err.Error()}
	}

	if viper.GetBool(config.RequireEmailForUser) && req.Email == nil {
		return models.User{}, svcErr.ValidationError{Message: "email is required"}
	}

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, fmt.Errorf("register user: %w", err)
	}

	user := models.User{
		ID:        uuid.Must(uuid.NewV7()), // V7 is sortable to make life easier for dear Postgres (faster INSERT and better index performance)
		Name:      req.Name,
		Username:  strings.ToLower(strings.TrimSpace(req.Username)),
		Email:     req.Email,
		Password:  string(hashedPasswordBytes),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err = svc.userStorage.CreateUser(ctx, user); err != nil {
		return models.User{}, svcErr.ReturnMappedStorageError(err, "create", "user", "ID", user.ID.String())
	}

	return user.RemovePassword(), nil
}

func (svc *UserServiceImpl) VerifyEmail(ctx context.Context, token string) error {
	userIDStr, err := svc.tokenStorage.GetEmailVerificationToken(ctx, token)
	if err != nil {
		if errors.Is(err, cacheErr.ErrEmailVerificationTokenNotFound) {
			return svcErr.ValidationError{Message: "invalid or expired verification token"}
		}
		return fmt.Errorf("verify email: get verification token: %w", err)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("verify email: failed to parse user ID: %w", err)
	}

	if err = svc.userStorage.SetUserEmailAsVerified(ctx, userID); err != nil {
		return svcErr.ReturnMappedStorageError(err, "verify email", "user", "ID", userID.String())
	}

	if err = svc.tokenStorage.DeleteEmailVerificationToken(ctx, token); err != nil {
		logger.ErrorWithID(ctx, "failed to delete email verification token for user %s: %v", userID, err)
	}

	return nil
}

func (svc *UserServiceImpl) LogIn(ctx context.Context, req apiModels.LogInUserRequest) (models.User, error) {
	req.Username = strings.ToLower(strings.TrimSpace(req.Username))

	user, err := svc.userStorage.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, svcErr.UnauthorizedError{Message: "invalid credentials"}
		}
		return models.User{}, fmt.Errorf("login: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return models.User{}, svcErr.UnauthorizedError{Message: "invalid credentials"}
	}

	if viper.GetBool(config.RequireEmailVerification) && (user.Email == nil || !user.EmailVerified) {
		return models.User{}, svcErr.ValidationError{Message: "email not verified"}
	}

	return user.RemovePassword(), nil
}

func (svc *UserServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	if id == uuid.Nil {
		return models.User{}, svcErr.ValidationError{Message: "user ID is required"}
	}

	user, err := svc.userStorage.GetUserByID(ctx, id)
	if err != nil {
		return models.User{}, svcErr.ReturnMappedStorageError(err, "get", "user", "ID", id.String())
	}

	return user, nil
}

func (svc *UserServiceImpl) UpdateUserByID(ctx context.Context, id uuid.UUID, req apiModels.UpdateUserRequest) error {
	if id == uuid.Nil {
		return svcErr.ValidationError{Message: "user ID is required"}
	}

	if err := req.Validate(); err != nil {
		return svcErr.ValidationError{Message: "validation failed: " + err.Error()}
	}

	user := models.User{UpdatedAt: time.Now()}
	fields := []string{"UpdatedAt"}
	if req.Name != nil {
		user.Name = *req.Name
		fields = append(fields, "Name")
	}
	if req.Username != nil {
		user.Username = strings.ToLower(strings.TrimSpace(*req.Username))
		fields = append(fields, "Username")
	}
	if req.Email != nil {
		user.Email = new(strings.ToLower(strings.TrimSpace(*req.Email)))
		fields = append(fields, "Email")
	}

	if err := svc.userStorage.UpdateUserByID(ctx, id, user, fields); err != nil {
		return svcErr.ReturnMappedStorageError(err, "update", "user", "ID", id.String())
	}

	return nil
}

func (svc *UserServiceImpl) UpdateAvatar(ctx context.Context, id uuid.UUID, avatarURL string) error {
	if id == uuid.Nil {
		return svcErr.ValidationError{Message: "user ID is required"}
	}

	user := models.User{
		Avatar:    &avatarURL,
		UpdatedAt: time.Now(),
	}
	fields := []string{"Avatar", "UpdatedAt"}

	if err := svc.userStorage.UpdateUserByID(ctx, id, user, fields); err != nil {
		return svcErr.ReturnMappedStorageError(err, "update avatar", "user", "ID", id.String())
	}

	return nil
}

func (svc *UserServiceImpl) DeleteAvatar(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return svcErr.ValidationError{Message: "user ID is required"}
	}

	if err := svc.userStorage.RemoveAvatar(ctx, id); err != nil {
		return svcErr.ReturnMappedStorageError(err, "remove avatar", "user", "ID", id.String())
	}

	return nil
}

func (svc *UserServiceImpl) VerifyPassword(ctx context.Context, id uuid.UUID, password string) error {
	user, err := svc.userStorage.GetUserByID(ctx, id)
	if err != nil {
		return svcErr.ReturnMappedStorageError(err, "verify password", "user", "ID", id.String())
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return svcErr.ValidationError{Message: "wrong password"}
	}

	return nil
}

func (svc *UserServiceImpl) ChangePassword(ctx context.Context, id uuid.UUID, req apiModels.ChangePasswordRequest) error {
	if err := req.Validate(); err != nil {
		return svcErr.ValidationError{Message: "validation failed: " + err.Error()}
	}

	user, err := svc.userStorage.GetUserByID(ctx, id)
	if err != nil {
		return svcErr.ReturnMappedStorageError(err, "change password", "user", "ID", id.String())
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return svcErr.ValidationError{Message: "wrong current password"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("change password: failed to hash: %w", err)
	}

	if err = svc.userStorage.UpdatePassword(ctx, id, string(hash)); err != nil {
		return svcErr.ReturnMappedStorageError(err, "change password", "user", "ID", id.String())
	}

	return nil
}

func (svc *UserServiceImpl) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := svc.userStorage.GetUserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.ErrorWithID(ctx, "failed to get user by email %s: %v", email, err)
		}
		return nil // No user? Keep it silent
	}
	if user.Email == nil {
		return nil // Also sealed lips
	}

	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		return fmt.Errorf("request password reset: %w", err)
	}

	if err = svc.tokenStorage.SavePasswordResetToken(ctx, token, user.ID.String()); err != nil {
		return fmt.Errorf("request password reset: %w", err)
	}

	if err = svc.emailService.SendPasswordResetLetter(ctx, *user.Email, token); err != nil {
		return fmt.Errorf("request password reset: %w", err)
	}

	return nil
}

func (svc *UserServiceImpl) ResetPassword(ctx context.Context, token, newPassword string) error {
	userIDString, err := svc.tokenStorage.GetPasswordResetToken(ctx, token)
	if err != nil {
		return svcErr.ValidationError{Message: "invalid or expired password reset token"}
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return fmt.Errorf("reset password: failed to parse user ID: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("reset password: failed to hash: %w", err)
	}

	if err = svc.userStorage.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return svcErr.ReturnMappedStorageError(err, "reset password", "user", "ID", userID.String())
	}

	if err = svc.tokenStorage.DeletePasswordResetToken(ctx, token); err != nil {
		logger.ErrorWithID(ctx, "failed to delete password reset token for user %s: %v", userID, err)
	}

	return nil
}

func (svc *UserServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := svc.userStorage.DeleteUserByID(ctx, id); err != nil {
		return svcErr.ReturnMappedStorageError(err, "delete", "user", "ID", id.String())
	}

	return nil
}
