package api

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"go.uber.org/fx"

	"go-fx-template/docs"
	"go-fx-template/internal/api/controllers"
	"go-fx-template/internal/api/middlewares"
	"go-fx-template/internal/config"
	"go-fx-template/internal/logger"
	"go-fx-template/internal/utils/gin"
	"go-fx-template/static"
)

type API struct {
	engine   *gin.Engine
	webCtrl  *controllers.WebController
	userCtrl *controllers.UserController
	itemCtrl *controllers.ItemController
}

func NewAPI(e *gin.Engine, wc *controllers.WebController, uc *controllers.UserController, ic *controllers.ItemController) *API {
	return &API{
		engine:   e,
		webCtrl:  wc,
		userCtrl: uc,
		itemCtrl: ic,
	}
}

func NewEngine() *gin.Engine {
	if viper.GetBool(config.GinReleaseMode) {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	_ = engine.SetTrustedProxies(nil) // Can nil produce an error? Or can a robot write a symphony?
	engine.HandleMethodNotAllowed = true

	return engine
}

func NewHTTPServer(api *API) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf("%s:%s", viper.GetString(config.ApiHost), viper.GetString(config.ApiPort)),
		Handler: api.engine,
	}
}

func (api *API) loadStaticFiles() {
	api.engine.SetHTMLTemplate(template.Must(template.New("").ParseFS(
		static.TemplatesFS,
		"templates/*.gohtml",
	)))
	api.engine.StaticFS("/static", http.FS(static.PublicFS))
}

func (api *API) registerMiddlewares() {
	api.engine.Use(middlewares.RequestID())
	api.engine.Use(ginutils.LoggingMiddlewares()...)
	api.engine.Use(middlewares.CORS())
}

func (api *API) registerRoutes() {
	// Web
	api.webCtrl.RegisterRoutes()

	// API
	api.userCtrl.RegisterRoutes()
	api.itemCtrl.RegisterRoutes()

	// Swagger
	docs.SwaggerInfo.Host = viper.GetString(config.WebAppDomain)
	docs.SwaggerInfo.BasePath = viper.GetString(config.ApiBasePath)
	api.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func Run(lifecycle fx.Lifecycle, sd fx.Shutdowner, api *API, server *http.Server) {
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			api.loadStaticFiles()
			api.registerMiddlewares()
			api.registerRoutes()

			go func() { logger.Info("Listening on %s...", server.Addr) }()

			go func() {
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Error("Server stopped unexpectedly: %v", err)
					_ = sd.Shutdown(fx.ExitCode(1))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return logger.InfoAction("Shutting down HTTP server...", func() error {
				shutdownCtx, cancel := context.WithTimeout(ctx, viper.GetDuration(config.ApiShutdownTimeout))
				defer cancel()

				if err := server.Shutdown(shutdownCtx); err != nil {
					return fmt.Errorf("graceful HTTP shutdown: %w", err)
				}

				return nil
			})
		},
	})
}
