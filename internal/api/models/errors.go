package apiModels

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"go-fx-template/internal/logger"
	"go-fx-template/internal/services/errors"
)

type APIResponse struct {
	Message string `json:"message" example:"email verified"`
}

type BadRequestError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"invalid request payload"`
}

type UnauthorizedError struct {
	Code    int    `json:"code" example:"401"`
	Message string `json:"message" example:"access token is invalid or expired"`
}

type ForbiddenError struct {
	Code    int    `json:"code" example:"403"`
	Message string `json:"message" example:"operation is forbidden"`
}

type NotFoundError struct {
	Code    int    `json:"code" example:"404"`
	Message string `json:"message" example:"user not found"`
}

type ConflictError struct {
	Code    int    `json:"code" example:"409"`
	Message string `json:"message" example:"user with this username already exists"`
}

type APIError struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"Internal server error"`
	Details string `json:"details,omitempty" example:"Your request ID is 37m019ce87f-e4ec-713c-9ce2-cdff39a77b31"`
}

func Error(ctx *gin.Context, code int, message string) {
	ctx.AbortWithStatusJSON(code, APIError{
		Code:    code,
		Message: message,
	})
}

func ErrorWithDetails(ctx *gin.Context, code int, message, details string) {
	ctx.AbortWithStatusJSON(code, APIError{
		Code:    code,
		Message: message,
		Details: details,
	})
}

func fieldLabel(field string) string {
	switch field {
	case "Name":
		return "name"
	case "Username":
		return "username"
	case "Email":
		return "email"
	case "Password":
		return "password"
	case "OldPassword":
		return "current password"
	case "NewPassword":
		return "new password"
	case "CurrentPassword":
		return "password"
	case "RefreshToken":
		return "refresh token"
	case "Token":
		return "token"
	case "Title":
		return "title"
	case "Notes":
		return "notes"
	case "Link":
		return "link"
	case "Price":
		return "price"
	case "Currency":
		return "currency"
	default:
		return strings.ToLower(field)
	}
}

func RespondWithBindError(ctx *gin.Context, err error) {
	var errStr string

	var validationErrs validator.ValidationErrors
	var unmarshalTypeErr *json.UnmarshalTypeError
	var syntaxErr *json.SyntaxError

	switch {
	case errors.As(err, &validationErrs) && len(validationErrs) > 0:
		fieldErr := validationErrs[0]
		field := fieldLabel(fieldErr.Field())
		switch fieldErr.Tag() {
		case "required":
			errStr = fmt.Sprintf("%s is required", field)
		case "email":
			errStr = fmt.Sprintf("%s must be a valid email address", field)
		case "min":
			errStr = fmt.Sprintf("%s must be at least %s characters", field, fieldErr.Param())
		case "max":
			errStr = fmt.Sprintf("%s must be at most %s characters", field, fieldErr.Param())
		case "one" + "of":
			errStr = fmt.Sprintf("%s has invalid value", field)
		default:
			errStr = fmt.Sprintf("%s is invalid", field)
		}
	case errors.As(err, &unmarshalTypeErr) && unmarshalTypeErr.Field != "":
		errStr = fmt.Sprintf("%s has invalid type", fieldLabel(unmarshalTypeErr.Field))
	case errors.As(err, &syntaxErr):
		errStr = "invalid JSON payload"
	default:
		errStr = "invalid request payload"
	}

	Error(ctx, http.StatusBadRequest, errStr)
}

func RespondWithServiceError(ctx *gin.Context, err error) bool {
	var unauthorizedErr svcErr.UnauthorizedError
	var validationErr svcErr.ValidationError
	var conflictErr svcErr.ConflictError
	var forbiddenErr svcErr.ForbiddenError
	var notFoundErr svcErr.NotFoundError

	switch {
	case errors.As(err, &unauthorizedErr):
		Error(ctx, http.StatusUnauthorized, unauthorizedErr.Error())
		return true
	case errors.As(err, &validationErr):
		Error(ctx, http.StatusBadRequest, validationErr.Error())
		return true
	case errors.As(err, &conflictErr):
		Error(ctx, http.StatusConflict, conflictErr.Error())
		return true
	case errors.As(err, &forbiddenErr):
		Error(ctx, http.StatusForbidden, forbiddenErr.Error())
		return true
	case errors.As(err, &notFoundErr):
		Error(ctx, http.StatusNotFound, notFoundErr.Error())
		return true
	default:
		return false
	}
}

func RespondWithInternalError(ctx *gin.Context, details string) {
	logger.ErrorWithID(ctx, details)
	ErrorWithDetails(ctx, 500, "Internal server error", fmt.Sprintf("Your request ID is %s", ctx.GetString("request_id")))
}
