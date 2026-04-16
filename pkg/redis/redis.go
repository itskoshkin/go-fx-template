package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/logging"
	"go.uber.org/fx"

	"go-fx-template/internal/utils/text"
)

// Logger is the narrow logging interface required by this package — lets the
// consumer inject any compatible logger without forcing pkg/redis to depend
// on the app's logger package.
type Logger interface {
	Info(format string, args ...any)
	InfoAction(message string, action func() error) error
}

type Config struct {
	Addr     string
	Port     string
	Password string
	Database int
}

const connectTimeout = 2 * time.Second

func NewClient(lifecycle fx.Lifecycle, cfg Config, log Logger) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	fmt.Print("Connecting to Redis...")

	redis.SetLogger(&logging.VoidLogger{})
	rc := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.Database,
	})

	if _, err := rc.Ping(ctx).Result(); err != nil {
		_ = rc.Close()
		fmt.Println()
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	fmt.Println(text.Green("    Done."))

	lifecycle.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return log.InfoAction("Closing Redis connection...", func() error {
				if err := rc.Close(); err != nil {
					return fmt.Errorf("closing Redis connection: %w", err)
				}
				return nil
			})
		},
	})

	return rc, nil
}
