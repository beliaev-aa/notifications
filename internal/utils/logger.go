package utils

import (
	"github.com/sirupsen/logrus"
	"os"
)

const (
	// ServiceName название сервиса для идентификации логов
	ServiceName = "notifications"
)

// serviceHook добавляет поле service во все записи логгера.
type serviceHook struct {
	serviceName string
}

func (h *serviceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *serviceHook) Fire(entry *logrus.Entry) error {
	entry.Data["service"] = h.serviceName
	return nil
}

// InitLogger создает новый логгер с JSON-форматом и добавляет поле "service"
// Уровень логирования устанавливается из переменной окружения LOG_LEVEL
// Если переменная не задана, используется уровень "debug"
func InitLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	// Получаем уровень логирования из ENV или используем значение по умолчанию
	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "debug"
	}

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		logLevel = logrus.DebugLevel
	}

	logger.SetLevel(logLevel)
	logger.AddHook(&serviceHook{serviceName: ServiceName})

	logger.WithFields(logrus.Fields{
		"service":  ServiceName,
		"logLevel": logLevel.String(),
	}).Debug("Logger initialized")

	return logger
}
