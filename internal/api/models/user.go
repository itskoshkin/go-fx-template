package apiModels

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"go-fx-template/internal/models"
)

var emailRe = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$")

type RegisterUserRequest struct {
	Name     string  `json:"name" binding:"required" example:"Alice"`
	Username string  `json:"username" binding:"required" example:"user421"`
	Email    *string `json:"email" binding:"omitempty,email" example:"alice421@email.com"`
	Password string  `json:"password" binding:"required,min=8" example:"P4s5w0rd"`
}

func (req *RegisterUserRequest) Validate() error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(req.Username) == "" {
		return errors.New("username is required")
	}
	if req.Email != nil {
		if strings.TrimSpace(*req.Email) == "" {
			return errors.New("email must be not empty or omitted")
		}
		if !emailRe.MatchString(*req.Email) {
			return errors.New("email must be valid")
		}
	}
	if strings.TrimSpace(req.Password) == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 8 {
		return errors.New("password length must be at least 8 characters")
	}

	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	if req.Email != nil {
		*req.Email = strings.ToLower(strings.TrimSpace(*req.Email))
	}

	return nil
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required" example:"befe235a381306c34ac2e235a565e8d47febd88df9c100fe2cfeab5f7654db75"`
}

type LogInUserRequest struct {
	Username string `json:"username" binding:"required" example:"user421"`
	Password string `json:"password" binding:"required,min=8" example:"P4s5w0rd"`
}

func (req *LogInUserRequest) Validate() error {
	if strings.TrimSpace(req.Username) == "" {
		return errors.New("username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return errors.New("password is required")
	}

	return nil
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDE5Y2QzNDktZDE3Ni03NTYyLWIwM2ItMWRiMjIyM2I5YTAxIiwiaXNzIjoid2lzaGxpc3QuaXRza29zaGtpbi5ydSIsInN1YiI6IjAxOWNkMzQ5LWQxNzYtNzU2Mi1iMDNiLTFkYjIyMjNiOWEwMSIsImF1ZCI6WyJ3aXNobGlzdC5pdHNrb3Noa2luLnJ1Il0sImV4cCI6MTc3MzY3NjgzMywibmJmIjoxNzczMDcyMDMzLCJpYXQiOjE3NzMwNzIwMzMsImp0aSI6IjBiZDU4YWVkLWE1YjYtNDRjNi1iOTdkLWFhYjc5ZTNlZGMxNCJ9._WVHH8QfmA4qbzIyKJKJa6h2uM2jeoRtg8PLJFKSau4"`
}

type AuthResponse struct {
	AuthTokensResponse
	User UserResponse `json:"user"`
}

type AuthTokensResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDE5ZDgyY2MtNjJmOS03MzRmLThhOTEtOTYyN2FmYmIxZTFkIiwiaXNzIjoieW91ci1hcHAtbmFtZSIsInN1YiI6IjAxOWQ4MmNjLTYyZjktNzM0Zi04YTkxLTk2MjdhZmJiMWUxZCIsImF1ZCI6WyJ5b3VyLWRvbWFpbi5jb20iXSwiZXhwIjoxNzc2MTAyNTMzLCJuYmYiOjE3NzYwMTYxMzMsImlhdCI6MTc3NjAxNjEzMywianRpIjoiNzhkM2VhYjYtODVmNC00ZTZiLWFlZGYtY2MzM2YyY2JhY2QyIn0.55_EMmEToK4_uG1qKfkJ2CN8NAQ3U2yYJ5j5azkNhU8"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name" binding:"omitempty,max=255" example:"Alice"`
	Username *string `json:"username" binding:"omitempty,max=64" example:"user421"`
	Email    *string `json:"email" binding:"omitempty,email" example:"alice421@email.com"`
}

func (req *UpdateUserRequest) Validate() error {
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return errors.New("name must be not empty or omitted")
	}
	if req.Username != nil && strings.TrimSpace(*req.Username) == "" {
		return errors.New("username must be not empty or omitted")
	}
	if req.Email != nil {
		if strings.TrimSpace(*req.Email) == "" {
			return errors.New("email must be not empty or omitted")
		}
		if !emailRe.MatchString(*req.Email) {
			return errors.New("email must be valid")
		}
	}

	return nil
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=8" example:"OldP4s5w0rd"`
	NewPassword     string `json:"new_password" binding:"required,min=8" example:"NewP4s5w0rd"`
}

func (req *ChangePasswordRequest) Validate() error {
	if strings.TrimSpace(req.CurrentPassword) == "" {
		return errors.New("current password is required")
	}
	if strings.TrimSpace(req.NewPassword) == "" {
		return errors.New("new password is required")
	}
	if strings.TrimSpace(req.CurrentPassword) == strings.TrimSpace(req.NewPassword) {
		return errors.New("current and new passwords must not be equal")
	}

	return nil
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"alice421@email.com"`
}

type SetNewPasswordRequest struct {
	Token       string `json:"token" binding:"required" example:"3b5b0860ed1be5c0fe6b18db6615bd05046b09677aa514a6e46c232cbff1bf7a"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"Str0ng3rP4s5w0rd"`
}

type DeleteAccountRequest struct {
	CurrentPassword string `json:"password" binding:"required" example:"P4s5w0rd"`
}

type UserResponse struct {
	ID            uuid.UUID `json:"id" example:"019d82cc-62f9-734f-8a91-9627afbb1e1d"`
	Name          string    `json:"name" example:"Alice"`
	Username      string    `json:"username" example:"user421"`
	Email         *string   `json:"email" example:"alice421@email.com"`
	EmailVerified bool      `json:"email_verified" example:"true"`
	Avatar        *string   `json:"avatar" example:"http://localhost:9000/go-fx-template/avatars/019d82cc-62f9-734f-8a91-9627afbb1e1d/019d82d4-a959-7d00-ab7d-ae1bc8d1a384"`
	CreatedAt     time.Time `json:"created_at" example:"2026-04-04T00:00:00.000000+03:00"`
	UpdatedAt     time.Time `json:"updated_at" example:"2026-04-04T00:00:00.000000+03:00"`
}

func ToUserResponse(user models.User) UserResponse {
	return UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Username:      user.Username,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Avatar:        user.Avatar,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
