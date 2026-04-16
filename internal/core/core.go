package core

import (
	"go.uber.org/fx"

	"go-fx-template/internal/api"
	"go-fx-template/internal/api/controllers"
	"go-fx-template/internal/api/middlewares"
	"go-fx-template/internal/config"
	"go-fx-template/internal/logger"
	"go-fx-template/internal/logger/adapters"
	"go-fx-template/internal/repository/cache"
	"go-fx-template/internal/repository/object"
	"go-fx-template/internal/repository/storage"
	"go-fx-template/internal/services"
	"go-fx-template/pkg/minio"
	"go-fx-template/pkg/postgres"
	"go-fx-template/pkg/redis"
)

func Load() *fx.App {
	return fx.New(
		fxDefaults(),
		fx.Invoke(
			config.LoadConfig,
			logger.SetupLogger,
		),
		config.ProvideConfigs,
		adapters.ProvideLoggers,
		fx.Provide(
			postgres.NewInstance,
			redis.NewClient,
			minio.NewClient,
			storage.NewUserStorage,
			storage.NewItemStorage,
			cache.NewTokenStorage,
			objectStorage.NewMinioStorage,
			services.NewAuthService,
			services.NewUserService,
			services.NewItemService,
			services.NewImageService,
			services.NewEmailService,
			middlewares.NewMiddlewares,
			controllers.NewWebController,
			controllers.NewUserController,
			controllers.NewItemController,
			api.NewHTTPServer,
			api.NewEngine,
			api.NewAPI,
		),
		fx.Invoke(api.Run),
	)
}

func fxDefaults() fx.Option {
	return fx.Options(
		fx.StartTimeout(config.FxStartUpTimeout),
		fx.StopTimeout(config.FxShutDownTimeout),
		fx.WithLogger(adapters.FxLogger),
	)
}
