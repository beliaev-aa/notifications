package app

import (
	"github.com/beliaev-aa/notifications/internal/adapter/http"
	"github.com/beliaev-aa/notifications/internal/adapter/notification"
	"github.com/beliaev-aa/notifications/internal/adapter/notification/channel"
	"github.com/beliaev-aa/notifications/internal/adapter/youtrack"
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/beliaev-aa/notifications/internal/service"
	"github.com/sirupsen/logrus"
	nethttp "net/http"
	"time"
)

// App представляет основное приложение с композицией всех зависимостей
type App struct {
	httpServer port.HTTPServer
}

// NewApp создает новый экземпляр приложения с инициализированными зависимостями
func NewApp(cfg *config.Config, logger *logrus.Logger) *App {
	// Создаем и настраиваем отправитель уведомлений
	notificationSender := setupNotificationSender(cfg, logger)

	// Создаем сервис конфигурации проектов
	projectConfigService := service.NewProjectConfigService(cfg, logger)

	youtrackParser := youtrack.NewParser(projectConfigService)
	webhookService := service.NewWebhookService(notificationSender, youtrackParser, logger)

	// Создаем HTTP адаптер с зависимостью
	httpServer := http.NewServer(&cfg.HTTP, webhookService, logger)

	return &App{
		httpServer: httpServer,
	}
}

// setupNotificationSender создает и настраивает отправитель уведомлений с зарегистрированными каналами
func setupNotificationSender(cfg *config.Config, logger *logrus.Logger) port.NotificationSender {
	// Создаем отправитель уведомлений
	notificationSender := notification.NewSender(logger)

	// Регистрируем каналы отправки уведомлений
	notificationSender.RegisterChannel(channel.NewLoggerChannel(logger))

	// Регистрируем Telegram канал (используется для проектов с telegram в allowedChannels)
	// Telegram канал создается только если указан bot_token
	if cfg.Telegram.BotToken != "" {
		httpClient := &nethttp.Client{
			Timeout: time.Duration(cfg.Telegram.Timeout) * time.Second,
		}
		notificationSender.RegisterChannel(channel.NewTelegramChannel(cfg.Telegram, logger, httpClient))
	}

	// Регистрируем VK Teams канал (используется для проектов с vkteams в allowedChannels)
	// VK Teams канал создается только если указан bot_token
	if cfg.VKTeams.BotToken != "" {
		httpClient := &nethttp.Client{
			Timeout: time.Duration(cfg.VKTeams.Timeout) * time.Second,
		}
		notificationSender.RegisterChannel(channel.NewVKTeamsChannel(cfg.VKTeams, logger, httpClient))
	}

	return notificationSender
}

// Run запускает приложение
func (a *App) Run() error {
	return a.httpServer.Start()
}
