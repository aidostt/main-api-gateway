package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reservista.kz/internal/config"
	"reservista.kz/internal/delivery"
	"reservista.kz/internal/server"
	"reservista.kz/pkg/dialog"
	"reservista.kz/pkg/logger"
	auth "reservista.kz/pkg/manager"
	"reservista.kz/pkg/s3client"
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
	dial := dialog.NewDialog(cfg.Authority,
		fmt.Sprintf("%v:%v", cfg.Users.Host, cfg.Users.Port),
		fmt.Sprintf("%v:%v", cfg.Reservations.Host, cfg.Reservations.Port),
		fmt.Sprintf("%v:%v", cfg.QRs.Host, cfg.QRs.Port),
		fmt.Sprintf("%v:%v", cfg.Notifications.Host, cfg.Notifications.Port),
	)
	s3Client := s3client.NewS3Client(cfg.AWS.Region, fmt.Sprintf("http://%s:%s", cfg.HTTP.Host, cfg.HTTP.Port), cfg.AWS.AccessKey, cfg.AWS.PrivateKey, cfg.AWS.Bucket)
	tokenManager, err := auth.NewManager(cfg.JWT.SigningKey)
	if err != nil {
		logger.Error(err)
		return
	}
	handlers := delivery.NewHandler(
		delivery.Handler{
			CookieTTL:    cfg.Cookie.Ttl,
			Environment:  cfg.Environment,
			Dialog:       dial,
			TokenManager: tokenManager,
			HttpAddress:  cfg.HTTP.Host + ":" + cfg.HTTP.Port,
			S3Client:     s3Client,
			PageDefault:  cfg.Limiter.PageDefault,
			LimitDefault: cfg.Limiter.ElementLimiterDefault,
		})
	// HTTP Server
	httpServer := server.NewServer(cfg, handlers.Init())
	go func() {
		if err := httpServer.Run(); !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("error occurred while running http server: %s\n", err.Error())
		}
	}()
	logger.Infof("HTTP Server started at %v:%v", cfg.HTTP.Host, cfg.HTTP.Port)

	// Graceful Shutdown of HTTP Server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := httpServer.Stop(ctx); err != nil {
		logger.Errorf("failed to stop server: %v", err)
	}

}
