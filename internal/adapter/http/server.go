package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server представляет HTTP сервер с поддержкой graceful shutdown
type Server struct {
	cfg            *config.HTTPConfig
	logger         *logrus.Logger
	webhookService port.WebhookService
}

// NewServer создает новый экземпляр HTTP сервера
func NewServer(cfg *config.HTTPConfig, webhookService port.WebhookService, logger *logrus.Logger) *Server {
	return &Server{
		cfg:            cfg,
		webhookService: webhookService,
		logger:         logger,
	}
}

// Start запускает HTTP сервер и обрабатывает graceful shutdown
func (s *Server) Start() error {
	router := NewRouter(s.webhookService, s.logger)

	srv := &http.Server{
		Addr:         s.cfg.Addr,
		Handler:      router,
		ReadTimeout:  time.Duration(s.cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.cfg.WriteTimeout) * time.Second,
	}

	serverErrors := make(chan error, 1)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		s.logger.WithFields(logrus.Fields{
			"addr": srv.Addr,
		}).Info("Starting web-server")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.WithError(err).Error("Error starting web server")
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("failed to start server: %w", err)
	case sig := <-quit:
		s.logger.WithFields(logrus.Fields{
			"signal": sig.String(),
		}).Info("Shutting down http-server...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			s.logger.WithError(err).Warn("Graceful shutdown timeout exceeded, forcing shutdown")
		} else {
			s.logger.WithError(err).Error("Graceful shutdown failed")
		}
		return err
	}

	s.logger.Info("Http-server stopped")
	return nil
}
