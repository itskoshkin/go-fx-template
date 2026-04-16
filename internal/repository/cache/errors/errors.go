package cacheErr

import (
	"errors"
)

var (
	ErrEmailVerificationTokenNotFound = errors.New("email verification token not found")
	ErrRefreshTokenSessionNotFound    = errors.New("refresh token session not found")
	ErrRefreshTokenReuseDetected      = errors.New("refresh token reuse detected")
)
