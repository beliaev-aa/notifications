package http

import (
	"context"
	"errors"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Config содержит конфигурацию HTTP сервера
type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
}

// Server представляет HTTP сервер с поддержкой graceful shutdown
type Server struct {
	cfg            *Config
	webhookService port.WebhookService
	logger         *logrus.Logger
}

// NewServer создает новый экземпляр HTTP сервера
func NewServer(cfg *Config, webhookService port.WebhookService, logger *logrus.Logger) *Server {
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
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		s.logger.WithFields(logrus.Fields{
			"addr": srv.Addr,
		}).Info("Starting web-server")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.WithError(err).Fatal("Error starting web server")
		}
	}()

	<-quit

	s.logger.Info("Shutting down http-server...")

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Graceful shutdown failed")
		return err
	}

	s.logger.Info("Http-server stopped")
	return nil
}
