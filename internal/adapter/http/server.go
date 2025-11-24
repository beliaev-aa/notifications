package http

import (
	"context"
	"errors"
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Graceful shutdown failed")
		return err
	}

	s.logger.Info("Http-server stopped")
	return nil
}
