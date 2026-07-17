package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"supermarket-backend/config"
	"supermarket-backend/controllers"
	"supermarket-backend/middleware"
)

// @title Supermarket Backend API
// @version 1.0
// @description REST API Server for the Supermarket Backend System.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 1. Initialize complete dependency graph using Google Wire
	app, err := config.InitializeApp()
	if err != nil {
		log.Fatalf("Critical boot failure: failed to initialize application: %v", err)
	}

	app.Logger.Info("SupermarketBackend mono-service booting...",
		zap.String("port", app.Config.PORT),
		zap.String("app_env", app.Config.APP_ENV),
		zap.String("log_level", app.Config.LOG_LEVEL),
	)

	// 2. Set Gin mode according to the deployment environment (independent of log verbosity)
	if app.Config.APP_ENV == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	// 3. Initialize authentication middleware and token blacklist
	blacklist := config.NewRistrettoTokenBlacklist(app.Cache)
	authMW := middleware.NewAuthMiddleware(app.Config.JWT_SECRET, blacklist)

	var allowedOrigins []string
	if app.Config.ALLOWED_ORIGINS != "" {
		allowedOrigins = strings.Split(app.Config.ALLOWED_ORIGINS, ",")
	}

	// 4. Register all application endpoints & global middlewares
	controllers.RegisterRoutes(
		r,
		app.UserController,
		app.CategoryController,
		app.ClientController,
		app.ProductController,
		app.PromotionController,
		app.SaleController,
		app.Logger,
		authMW.Authenticate(),
		authMW.AuthorizeRoles,
		authMW.ValidateOwnerOrAdmin(),
		allowedOrigins,
	)

	app.Logger.Info("Gin router registered successfully. Starting server...")

	// 5. Run the Gin monolithic HTTP/REST server behind an http.Server with
	// hardened timeouts and a graceful shutdown path.
	srv := &http.Server{
		Addr:              ":" + app.Config.PORT,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}
		serverErrCh <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		app.Logger.Info("Shutdown signal received, draining in-flight requests...", zap.String("signal", sig.String()))
	case err := <-serverErrCh:
		if err != nil {
			app.Logger.Fatal("Critical server crash: failed to run HTTP server", zap.Error(err))
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		app.Logger.Error("Graceful shutdown failed, forcing close", zap.Error(err))
		_ = srv.Close()
	}

	if sqlDB, dbErr := app.DB.DB(); dbErr == nil {
		if err := sqlDB.Close(); err != nil {
			app.Logger.Error("Failed to close database connection pool", zap.Error(err))
		}
	}
	app.Cache.Close()

	app.Logger.Info("SupermarketBackend shut down cleanly.")
}
