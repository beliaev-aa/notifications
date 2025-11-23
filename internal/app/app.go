package app

import (
	"github.com/beliaev-aa/notifications/internal/adapter/http"
	"github.com/beliaev-aa/notifications/internal/adapter/notification"
	"github.com/beliaev-aa/notifications/internal/adapter/notification/channel"
	"github.com/beliaev-aa/notifications/internal/service"
	"github.com/sirupsen/logrus"
	"time"
)

// Config содержит конфигурацию приложения
type Config struct {
	HTTPAddr        string
	ShutdownTimeout time.Duration
}

// App представляет основное приложение с композицией всех зависимостей
type App struct {
	httpServer *http.Server
}

// NewApp создает новый экземпляр приложения с инициализированными зависимостями
func NewApp(cfg *Config, logger *logrus.Logger) *App {
	// Создаем отправитель уведомлений
	notificationSender := notification.NewSender(logger)

	// Регистрируем каналы отправки уведомлений
	notificationSender.RegisterChannel(channel.NewLoggerChannel(logger))

	webhookService := service.NewWebhookService(notificationSender, logger)

	// Создаем HTTP конфигурацию адаптера
	httpCfg := &http.Config{
		Addr:            cfg.HTTPAddr,
		ShutdownTimeout: cfg.ShutdownTimeout,
	}

	// Создаем HTTP адаптер с зависимостью
	httpServer := http.NewServer(httpCfg, webhookService, logger)

	return &App{
		httpServer: httpServer,
	}
}

// Run запускает приложение
func (a *App) Run() error {
	return a.httpServer.Start()
}
