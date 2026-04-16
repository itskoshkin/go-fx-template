package adapters

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	gormlogger "gorm.io/gorm/logger"

	"go-fx-template/internal/config"
	"go-fx-template/internal/logger"
	gormutils "go-fx-template/internal/utils/gorm"
	"go-fx-template/pkg/postgres"
	"go-fx-template/pkg/redis"
)

func FxLogger() fxevent.Logger   { return CustomFxLogger{} }
func FxNoLogger() fxevent.Logger { return fxevent.NopLogger }

func PostgresLogger() postgres.Logger { return logger.GlobalLogger{} }

func RedisLogger() redis.Logger { return logger.GlobalLogger{} }

func GormLogger(lc fx.Lifecycle) gormlogger.Interface {
	level, showRuntime := gormutils.RuntimeLoggerConfig(viper.GetString(config.GormLogLevel))
	l := gormutils.NewCustomLogger(level, showRuntime, viper.GetBool(config.GormStartupLogs))

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			gormutils.EnableTimestamps()
			return nil
		},
	})

	return l
}

var ProvideLoggers = fx.Module("adapters",
	fx.Provide(PostgresLogger, RedisLogger, GormLogger),
)
