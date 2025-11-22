package app

import (
	"github.com/beliaev-aa/notifications/internal/adapter/http"
	"github.com/beliaev-aa/notifications/internal/usecase"
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
func NewApp(cfg *Config, log *logrus.Logger) *App {
	webhookUseCase := usecase.NewWebhookUseCase(log)

	// Создаем HTTP конфигурацию адаптера
	httpCfg := &http.Config{
		Addr:            cfg.HTTPAddr,
		ShutdownTimeout: cfg.ShutdownTimeout,
	}

	// Создаем HTTP адаптер с зависимостью
	httpServer := http.NewServer(httpCfg, webhookUseCase, log)

	return &App{
		httpServer: httpServer,
	}
}

// Run запускает приложение
func (a *App) Run() error {
	return a.httpServer.Start()
}
