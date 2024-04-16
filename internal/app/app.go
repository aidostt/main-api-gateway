package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"reservista.kz/internal/config"
	"reservista.kz/internal/delivery"
	"reservista.kz/internal/repository"
	"reservista.kz/internal/server"
	"reservista.kz/internal/service"
	"reservista.kz/pkg/database/mongodb"
	"reservista.kz/pkg/hash"
	"reservista.kz/pkg/logger"
	auth "reservista.kz/pkg/manager"
	"syscall"
	"time"
)

func Run(configPath, envPath string) {
	cfg, err := config.Init(configPath, envPath)
	if err != nil {
		logger.Error(err)

		return
	}

	// Dependencies
	mongoClient, err := mongodb.NewClient(cfg.Mongo.URI, cfg.Mongo.User, cfg.Mongo.Password)
	if err != nil {
		logger.Error(err)

		return
	}
	db := mongoClient.Database(cfg.Mongo.Name)

	hasher := hash.NewSHA1Hasher(cfg.Auth.PasswordSalt)

	tokenManager, err := auth.NewManager(cfg.Auth.JWT.SigningKey)
	if err != nil {
		logger.Error(err)

		return
	}

	repos := repository.NewModels(db)
	services := service.NewServices(service.Dependencies{
		Repos:           repos,
		Hasher:          hasher,
		TokenManager:    tokenManager,
		AccessTokenTTL:  cfg.Auth.JWT.AccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.JWT.RefreshTokenTTL,
		Environment:     cfg.Environment,
		Domain:          cfg.HTTP.Host,
	})
	handlers := delivery.NewHandler(services, tokenManager)

	// HTTP Server
	srv := server.NewServer(cfg, handlers.Init())

	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("error occurred while running http server: %s\n", err.Error())
		}
	}()

	logger.Info("Server started")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		logger.Errorf("failed to stop server: %v", err)
	}

	if err := mongoClient.Disconnect(context.Background()); err != nil {
		logger.Error(err.Error())
	}
}
