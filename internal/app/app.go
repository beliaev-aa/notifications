package app

import (
	"github.com/beliaev-aa/notifications/internal/adapter/http"
	"github.com/beliaev-aa/notifications/internal/adapter/notification"
	"github.com/beliaev-aa/notifications/internal/adapter/notification/channel"
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/beliaev-aa/notifications/internal/service"
	"github.com/sirupsen/logrus"
)

// App представляет основное приложение с композицией всех зависимостей
type App struct {
	httpServer port.HTTPServer
}

// NewApp создает новый экземпляр приложения с инициализированными зависимостями
func NewApp(cfg *config.Config, logger *logrus.Logger) *App {
	// Создаем отправитель уведомлений
	notificationSender := notification.NewSender(logger)

	// Регистрируем каналы отправки уведомлений
	notificationSender.RegisterChannel(channel.NewLoggerChannel(logger))

	// Регистрируем Telegram канал, если он включен
	if cfg.Telegram.Enabled {
		notificationSender.RegisterChannel(channel.NewTelegramChannel(cfg.Telegram, logger))
	}

	webhookService := service.NewWebhookService(notificationSender, logger)

	// Создаем HTTP адаптер с зависимостью
	httpServer := http.NewServer(&cfg.HTTP, webhookService, logger)

	return &App{
		httpServer: httpServer,
	}
}

// Run запускает приложение
func (a *App) Run() error {
	return a.httpServer.Start()
}
