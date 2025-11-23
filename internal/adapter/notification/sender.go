package notification

import (
	"fmt"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/sirupsen/logrus"
)

// Sender реализует порт NotificationSender для отправки уведомлений через различные каналы
type Sender struct {
	channels map[string]port.NotificationChannel
	logger   *logrus.Logger
}

// NewSender создает новый экземпляр отправителя уведомлений
func NewSender(logger *logrus.Logger) port.NotificationSender {
	return &Sender{
		channels: make(map[string]port.NotificationChannel),
		logger:   logger,
	}
}

// Send отправляет уведомление через указанный канал
func (s *Sender) Send(channel string, formattedMessage string) error {
	if formattedMessage == "" {
		return fmt.Errorf("formatted message cannot be empty")
	}

	ch, exists := s.channels[channel]
	if !exists {
		return fmt.Errorf("channel '%s' is not registered", channel)
	}

	return ch.Send(formattedMessage)
}

// RegisterChannel регистрирует новый канал для отправки уведомлений
func (s *Sender) RegisterChannel(channel port.NotificationChannel) {
	if channel == nil {
		s.logger.Warn("Attempted to register nil channel")
		return
	}

	channelName := channel.Channel()
	if channelName == "" {
		s.logger.Warn("Attempted to register channel with empty name")
		return
	}

	s.channels[channelName] = channel
	s.logger.WithField("channel", channelName).Info("Notification channel registered")
}
