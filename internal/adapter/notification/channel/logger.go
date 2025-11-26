package channel

import (
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/sirupsen/logrus"
)

// LoggerChannel реализует канал отправки уведомлений через логирование (для тестирования и разработки)
type LoggerChannel struct {
	logger *logrus.Logger
}

// NewLoggerChannel создает новый канал логирования уведомлений
func NewLoggerChannel(logger *logrus.Logger) port.NotificationChannel {
	return &LoggerChannel{
		logger: logger,
	}
}

// Send отправляет уведомление в логи
func (c *LoggerChannel) Send(_ string, formattedMessage string) error {
	c.logger.WithFields(logrus.Fields{
		"message": formattedMessage,
	}).Info("Notification sent via logger channel")

	return nil
}

// Channel возвращает название канала
func (c *LoggerChannel) Channel() string {
	return port.ChannelLogger
}
