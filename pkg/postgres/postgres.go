package postgres

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"go-fx-template/internal/models"
	"go-fx-template/internal/utils/text"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	ConnectionTimeout time.Duration

	AutoMigrate bool

	StartupLogs bool
}

type Logger interface {
	Info(format string, args ...any)
	InfoWithAction(message string, action func() error) error
}

func NewInstance(lifecycle fx.Lifecycle, cfg Config, log Logger, gormLogger gormlogger.Interface) (*gorm.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectionTimeout)
	defer cancel()

	fmt.Print("Connecting to Postgres...")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)

	if cfg.StartupLogs {
		fmt.Println()
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLogger})
	if err != nil {
		fmt.Println()
		return nil, fmt.Errorf("GORM: failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println()
		return nil, fmt.Errorf("GORM: failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err = sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		fmt.Println()
		return nil, fmt.Errorf("GORM: failed to ping database: %w", err)
	}

	if cfg.AutoMigrate {
		if err = db.AutoMigrate(&models.User{}, &models.Item{}); err != nil {
			_ = sqlDB.Close()
			fmt.Println()
			return nil, fmt.Errorf("GORM: failed to auto-migrate: %w", err)
		}
	}

	if cfg.StartupLogs {
		fmt.Println(text.Green("Done."))
	} else {
		fmt.Println(text.Green(" Done."))
	}

	lifecycle.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return log.InfoWithAction("Closing Postgres connection...", func() error {
				if err = sqlDB.Close(); err != nil {
					return fmt.Errorf("closing Postgres connection: %w", err)
				}
				return nil
			})
		},
	})

	return db, nil
}
