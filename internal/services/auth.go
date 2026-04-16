package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"

	"go-fx-template/internal/api/controllers"
	"go-fx-template/internal/api/middlewares"
	"go-fx-template/internal/config"
	"go-fx-template/internal/repository/cache/errors"
	"go-fx-template/internal/services/errors"
)

type TokenStorage interface {
	SaveEmailVerificationToken(ctx context.Context, tokenID, userID string) error
	GetEmailVerificationToken(ctx context.Context, tokenID string) (string, error)
	DeleteEmailVerificationToken(ctx context.Context, tokenID string) error
	CheckIfAuthTokenRevoked(ctx context.Context, tokenID string) (bool, error)
	RevokeAuthTokens(ctx context.Context, tokenID string, remainingTTL time.Duration) error
	SaveRefreshTokenSession(ctx context.Context, tokenID, familyID string, ttl time.Duration) error
	GetRefreshTokenSession(ctx context.Context, tokenID string) (familyID, status string, err error)
	RotateRefreshTokenSession(ctx context.Context, oldTokenID, newTokenID, familyID string, newTTL time.Duration) error
	CheckIfRefreshTokenFamilyRevoked(ctx context.Context, familyID string) (bool, error)
	RevokeRefreshTokenFamily(ctx context.Context, familyID string, ttl time.Duration) error
	SavePasswordResetToken(ctx context.Context, tokenID string, userID string) error
	GetPasswordResetToken(ctx context.Context, tokenID string) (string, error)
	DeletePasswordResetToken(ctx context.Context, tokenID string) error
}

type AuthServiceImpl struct {
	tokenStorage       TokenStorage
	accessTokenSecret  string
	refreshTokenSecret string
	accessTokenTTL     time.Duration
	refreshTokenTTL    time.Duration
}

func newAuthService(ts TokenStorage) *AuthServiceImpl {
	return &AuthServiceImpl{
		tokenStorage:       ts,
		accessTokenSecret:  viper.GetString(config.AccessTokenSecret),
		refreshTokenSecret: viper.GetString(config.RefreshTokenSecret),
		accessTokenTTL:     viper.GetDuration(config.AccessTokenTTL),
		refreshTokenTTL:    viper.GetDuration(config.RefreshTokenTTL),
	}
}

func NewAuthService(ts TokenStorage) (controllers.AuthService, middlewares.AuthService) {
	authService := newAuthService(ts)
	return authService, authService
} // Not the best idea TBH...

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func issueNewToken(userID uuid.UUID, ttl time.Duration) (*jwt.Token, Claims) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   userID.String(),
			Issuer:    viper.GetString(config.JwtIssuer),
			Audience:  viper.GetStringSlice(config.JwtAudience),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims), claims
}

func (svc *AuthServiceImpl) validateToken(tokenString, secretString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretString), nil
	}, jwt.WithAudience(viper.GetStringSlice(config.JwtAudience)...))
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (svc *AuthServiceImpl) validateActiveRefreshToken(ctx context.Context, tokenString string) (*Claims, string, error) {
	claims, err := svc.validateToken(tokenString, svc.refreshTokenSecret)
	if err != nil {
		return nil, "", svcErr.ErrRefreshTokenInvalidOrExpired
	}

	revoked, err := svc.tokenStorage.CheckIfAuthTokenRevoked(ctx, claims.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check token revocation: %w", err)
	}

	if revoked {
		return nil, "", svcErr.ErrRefreshTokenInvalidOrExpired
	}

	familyID, status, err := svc.tokenStorage.GetRefreshTokenSession(ctx, claims.ID)
	if err != nil {
		if errors.Is(err, cacheErr.ErrRefreshTokenSessionNotFound) {
			return nil, "", svcErr.ErrRefreshTokenInvalidOrExpired
		}
		return nil, "", fmt.Errorf("failed to get refresh token session: %w", err)
	}

	familyRevoked, err := svc.tokenStorage.CheckIfRefreshTokenFamilyRevoked(ctx, familyID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check refresh token family revocation: %w", err)
	}

	if familyRevoked {
		return nil, "", svcErr.ErrRefreshTokenCompromised
	}

	if status != "active" {
		if revokeErr := svc.tokenStorage.RevokeRefreshTokenFamily(ctx, familyID, svc.refreshTokenTTL); revokeErr != nil {
			return nil, "", fmt.Errorf("failed to revoke refresh token family after reuse detection: %w", revokeErr)
		}

		return nil, "", svcErr.ErrRefreshTokenCompromised
	}

	return claims, familyID, nil
}

func (svc *AuthServiceImpl) GenerateTokens(ctx context.Context, userID uuid.UUID) (access, refresh string, err error) {
	accessToken, _ := issueNewToken(userID, svc.accessTokenTTL)
	signedAccessToken, err := accessToken.SignedString([]byte(svc.accessTokenSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken, refreshClaims := issueNewToken(userID, svc.refreshTokenTTL)
	signedRefreshToken, err := refreshToken.SignedString([]byte(svc.refreshTokenSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	if err = svc.tokenStorage.SaveRefreshTokenSession(ctx, refreshClaims.ID, refreshClaims.ID, svc.refreshTokenTTL); err != nil {
		return "", "", fmt.Errorf("failed to persist refresh token session: %w", err)
	}

	return signedAccessToken, signedRefreshToken, nil
}

func (svc *AuthServiceImpl) ValidateAccessToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	claims, err := svc.validateToken(tokenString, svc.accessTokenSecret)
	if err != nil {
		return uuid.Nil, err
	}

	revoked, err := svc.tokenStorage.CheckIfAuthTokenRevoked(ctx, claims.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to check token revocation: %w", err)
	}

	if revoked {
		return uuid.Nil, fmt.Errorf("token has been revoked")
	}

	return claims.UserID, nil
}

func (svc *AuthServiceImpl) ValidateRefreshToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	claims, _, err := svc.validateActiveRefreshToken(ctx, tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	return claims.UserID, nil
}

func (svc *AuthServiceImpl) RotateRefreshToken(ctx context.Context, refreshToken string, userID uuid.UUID) (access, nextRefresh string, err error) {
	claims, familyID, err := svc.validateActiveRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	if claims.UserID != userID {
		return "", "", svcErr.ErrRefreshTokenInvalidOrExpired
	}

	accessToken, _ := issueNewToken(claims.UserID, svc.accessTokenTTL)
	signedAccessToken, err := accessToken.SignedString([]byte(svc.accessTokenSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	newRefreshToken, newRefreshClaims := issueNewToken(claims.UserID, svc.refreshTokenTTL)
	signedRefreshToken, err := newRefreshToken.SignedString([]byte(svc.refreshTokenSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	if err = svc.tokenStorage.RotateRefreshTokenSession(ctx, claims.ID, newRefreshClaims.ID, familyID, svc.refreshTokenTTL); err != nil {
		if errors.Is(err, cacheErr.ErrRefreshTokenReuseDetected) {
			if revokeErr := svc.tokenStorage.RevokeRefreshTokenFamily(ctx, familyID, svc.refreshTokenTTL); revokeErr != nil {
				return "", "", fmt.Errorf("failed to revoke refresh token family after reuse detection: %w", revokeErr)
			}

			return "", "", svcErr.ErrRefreshTokenCompromised
		}

		if errors.Is(err, cacheErr.ErrRefreshTokenSessionNotFound) {
			return "", "", svcErr.ErrRefreshTokenInvalidOrExpired
		}

		return "", "", fmt.Errorf("failed to rotate refresh token session: %w", err)
	}

	return signedAccessToken, signedRefreshToken, nil
}

func (svc *AuthServiceImpl) RevokeAuthTokens(ctx context.Context, accessToken, refreshToken string) error {
	errs := make([]error, 0)

	if claims, err := svc.validateToken(accessToken, svc.accessTokenSecret); err == nil {
		if remaining := time.Until(claims.ExpiresAt.Time); remaining > 0 {
			if err = svc.tokenStorage.RevokeAuthTokens(ctx, claims.ID, remaining); err != nil {
				errs = append(errs, fmt.Errorf("access token: %w", err))
			}
		}
	}

	if claims, err := svc.validateToken(refreshToken, svc.refreshTokenSecret); err == nil {
		if remaining := time.Until(claims.ExpiresAt.Time); remaining > 0 {
			if err = svc.tokenStorage.RevokeAuthTokens(ctx, claims.ID, remaining); err != nil {
				errs = append(errs, fmt.Errorf("refresh token: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to revoke tokens: %v", errs)
	}

	return nil
}
