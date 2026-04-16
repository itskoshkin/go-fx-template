package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"

	"go-fx-template/internal/config"
	"go-fx-template/internal/repository/cache/errors"
	"go-fx-template/internal/services"
)

const (
	projectPrefix                   = "go-fx-template"
	emailVerificationPrefix         = projectPrefix + ":" + "email_verification_token:"
	passwordResetPrefix             = projectPrefix + ":" + "password_reset_token:"
	refreshTokenSessionPrefix       = projectPrefix + ":" + "refresh_token_session:"
	revokedAuthTokensPrefix         = projectPrefix + ":" + "revoked_auth_token:"
	refreshTokenFamilyRevokedPrefix = projectPrefix + ":" + "refresh_token_family_revoked:"
	refreshTokenStatusActive        = "active"
	refreshTokenStatusRotated       = "rotated"
)

type refreshTokenSession struct {
	FamilyID string `json:"family_id"`
	Status   string `json:"status"`
}

type TokenStorageImpl struct {
	client                    *redis.Client
	passwordResetTokenTTL     time.Duration
	emailVerificationTokenTTL time.Duration
}

func newTokenStorage(client *redis.Client) *TokenStorageImpl {
	return &TokenStorageImpl{client: client, passwordResetTokenTTL: viper.GetDuration(config.PasswordResetTokenTTL), emailVerificationTokenTTL: viper.GetDuration(config.EmailVerificationTokenTTL)}
}

func NewTokenStorage(client *redis.Client) services.TokenStorage { return newTokenStorage(client) }

func (ts *TokenStorageImpl) SaveEmailVerificationToken(ctx context.Context, tokenID, userID string) error {
	return ts.client.Set(ctx, emailVerificationPrefix+tokenID, userID, ts.emailVerificationTokenTTL).Err()
}

func (ts *TokenStorageImpl) GetEmailVerificationToken(ctx context.Context, tokenID string) (string, error) {
	userID, err := ts.client.Get(ctx, emailVerificationPrefix+tokenID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", cacheErr.ErrEmailVerificationTokenNotFound
		}
		return "", err
	}

	return userID, nil
}

func (ts *TokenStorageImpl) DeleteEmailVerificationToken(ctx context.Context, tokenID string) error {
	return ts.client.Del(ctx, emailVerificationPrefix+tokenID).Err()
}

func (ts *TokenStorageImpl) CheckIfAuthTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	if err := ts.client.Get(ctx, revokedAuthTokensPrefix+tokenID).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (ts *TokenStorageImpl) RevokeAuthTokens(ctx context.Context, tokenID string, remainingTTL time.Duration) error {
	return ts.client.Set(ctx, revokedAuthTokensPrefix+tokenID, "ACTIVE, REVOKED", remainingTTL).Err()
}

func (ts *TokenStorageImpl) SaveRefreshTokenSession(ctx context.Context, tokenID, familyID string, ttl time.Duration) error {
	payload, err := marshalRefreshTokenSession(refreshTokenSession{FamilyID: familyID, Status: refreshTokenStatusActive})
	if err != nil {
		return err
	}

	return ts.client.Set(ctx, refreshTokenSessionPrefix+tokenID, payload, ttl).Err()
}

func (ts *TokenStorageImpl) GetRefreshTokenSession(ctx context.Context, tokenID string) (familyID, status string, err error) {
	payload, err := ts.client.Get(ctx, refreshTokenSessionPrefix+tokenID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", "", cacheErr.ErrRefreshTokenSessionNotFound
		}
		return "", "", err
	}

	record, err := unmarshalRefreshTokenSession(payload)
	if err != nil {
		return "", "", err
	}

	return record.FamilyID, record.Status, nil
}

func (ts *TokenStorageImpl) RotateRefreshTokenSession(ctx context.Context, oldTokenID, newTokenID, familyID string, newTTL time.Duration) error {
	return ts.client.Watch(ctx, func(tx *redis.Tx) error {
		payload, err := tx.Get(ctx, refreshTokenSessionPrefix+oldTokenID).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return cacheErr.ErrRefreshTokenSessionNotFound
			}
			return err
		}

		record, err := unmarshalRefreshTokenSession(payload)
		if err != nil {
			return err
		}

		if record.Status != refreshTokenStatusActive {
			return cacheErr.ErrRefreshTokenReuseDetected
		}

		rotatedPayload, err := marshalRefreshTokenSession(refreshTokenSession{FamilyID: familyID, Status: refreshTokenStatusRotated})
		if err != nil {
			return err
		}

		activePayload, err := marshalRefreshTokenSession(refreshTokenSession{FamilyID: familyID, Status: refreshTokenStatusActive})
		if err != nil {
			return err
		}

		pipe := tx.TxPipeline()
		pipe.SetArgs(ctx, refreshTokenSessionPrefix+oldTokenID, rotatedPayload, redis.SetArgs{KeepTTL: true})
		pipe.Set(ctx, refreshTokenSessionPrefix+newTokenID, activePayload, newTTL)

		if _, err = pipe.Exec(ctx); err != nil {
			if errors.Is(err, redis.TxFailedErr) {
				return cacheErr.ErrRefreshTokenReuseDetected
			}
			return err
		}

		return nil
	}, refreshTokenSessionPrefix+oldTokenID)
}

func (ts *TokenStorageImpl) CheckIfRefreshTokenFamilyRevoked(ctx context.Context, familyID string) (bool, error) {
	if err := ts.client.Get(ctx, refreshTokenFamilyRevokedPrefix+familyID).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (ts *TokenStorageImpl) RevokeRefreshTokenFamily(ctx context.Context, familyID string, ttl time.Duration) error {
	return ts.client.Set(ctx, refreshTokenFamilyRevokedPrefix+familyID, "REVOKED", ttl).Err()
}

func (ts *TokenStorageImpl) SavePasswordResetToken(ctx context.Context, tokenID string, userID string) error {
	return ts.client.Set(ctx, passwordResetPrefix+tokenID, userID, ts.passwordResetTokenTTL).Err()
}

func (ts *TokenStorageImpl) GetPasswordResetToken(ctx context.Context, tokenID string) (string, error) {
	return ts.client.Get(ctx, passwordResetPrefix+tokenID).Result()
}

func (ts *TokenStorageImpl) DeletePasswordResetToken(ctx context.Context, tokenID string) error {
	return ts.client.Del(ctx, passwordResetPrefix+tokenID).Err()
}

func marshalRefreshTokenSession(record refreshTokenSession) (string, error) {
	payload, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("failed to marshal refresh token session: %w", err)
	}

	return string(payload), nil
}

func unmarshalRefreshTokenSession(payload string) (refreshTokenSession, error) {
	var record refreshTokenSession
	if err := json.Unmarshal([]byte(payload), &record); err != nil {
		return refreshTokenSession{}, fmt.Errorf("failed to unmarshal refresh token session: %w", err)
	}

	return record, nil
}
