package redisrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type TokenRepository struct {
	client *redis.Client
}

func NewTokenRepository(client *redis.Client) *TokenRepository {
	return &TokenRepository{client: client}
}

func (r *TokenRepository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenID string, expiresIn time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", userID.String())

	data := map[string]interface{}{
		"token_id": tokenID,
		"user_id":  userID.String(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, jsonData, expiresIn).Err()
}

func (r *TokenRepository) GetRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", userID.String())

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}

	var tokenData map[string]string
	if err := json.Unmarshal([]byte(data), &tokenData); err != nil {
		return "", err
	}

	return tokenData["token_id"], nil
}

func (r *TokenRepository) DeleteRefreshToken(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf("refresh_token:%s", userID.String())
	return r.client.Del(ctx, key).Err()
}

func (r *TokenRepository) StoreBlacklistedToken(ctx context.Context, token string, expiresIn time.Duration) error {
	key := fmt.Sprintf("blacklisted_token:%s", token)
	return r.client.Set(ctx, key, "1", expiresIn).Err()
}

func (r *TokenRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklisted_token:%s", token)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}
