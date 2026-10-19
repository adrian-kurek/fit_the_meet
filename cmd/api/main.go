package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"

	middleware "github.com/adrian-kurek/fit_the_meet/common/authentication"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/logger"
	"github.com/adrian-kurek/fit_the_meet/config"
	"github.com/adrian-kurek/fit_the_meet/internal/auth"
	"github.com/adrian-kurek/fit_the_meet/internal/server"
	"github.com/adrian-kurek/fit_the_meet/internal/user"
	"github.com/joho/godotenv"
)

func bootstrapHTTPServer(
	db *sql.DB,
	cache commoninterfaces.CacheService,
	port string,
	logger commoninterfaces.Logger,
) *server.HTTP {
	authRepository := auth.NewRepository(db, logger)
	userRepository := user.NewRepository(db, logger)
	authenticationMiddleware := middleware.NewMiddleware(
		os.Getenv("ACCESS_TOKEN_SECRET"),
		os.Getenv("REFRESH_TOKEN_SECRET"),
		logger,
		cache,
	)
	authService := auth.NewService(authRepository, userRepository, authenticationMiddleware, logger)
	authHandler := auth.NewHandler(authService, authenticationMiddleware, logger)
	dependenciesConfig := server.NewDependencyConfig(
		port,
		logger,
		authHandler,
	)
	return server.NewHTTP(dependenciesConfig)
}

func main() {
	logger := logger.Setup()
	err := godotenv.Load()
	if err != nil {
		logger.Error("failed to load environment variables", "err", err.Error())
		panic(err)
	}

	db, err := config.ConnectToDB()
	if err != nil {
		logger.Error("failed to connect to databse service", "err", err.Error())
		panic(err)
	}

	defer db.Close()

	cacheService, err := config.ConnectToCache()
	if err != nil {
		logger.Error("failed to connect to cache service", "err", err.Error())
		panic(err)
	}
	defer cacheService.Close()

	port := os.Getenv("PORT")

	httpServer := bootstrapHTTPServer(db.Connection, cacheService, port, logger)
	apiCtx, apiCtxCancel := context.WithCancel(context.Background())
	go func() {
		logger.Info("server started", "port", port)
		if err = httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Failed to start server", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	defer apiCtxCancel()

	if err = httpServer.Shutdown(apiCtx); err != nil {
		logger.Error("Server forced to shutdown", "err", err)
		if err != nil {
			panic(err)
		}
		err = db.Close()
		if err != nil {
			panic(err)
		}
		err = cacheService.Close()
		if err != nil {
			panic(err)
		}
	}

	logger.Info("server exited")
}
