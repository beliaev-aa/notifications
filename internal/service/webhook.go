package service

import (
	"github.com/beliaev-aa/notifications/internal/adapter/youtrack"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// WebhookService реализует бизнес-логику обработки webhook запросов
type WebhookService struct {
	notificationSender port.NotificationSender
	youtrackParser     parser.YoutrackParser
	logger             *logrus.Logger
}

// NewWebhookService создает новый экземпляр сервиса для обработки webhook запросов
func NewWebhookService(notificationSender port.NotificationSender, logger *logrus.Logger) port.WebhookService {
	return &WebhookService{
		notificationSender: notificationSender,
		youtrackParser:     youtrack.NewParser(),
		logger:             logger,
	}
}

// ProcessWebhook обрабатывает входящий webhook запрос: читает тело, декодирует JSON и отправляет уведомление
func (w *WebhookService) ProcessWebhook(req *http.Request) error {
	// Читаем тело запроса
	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.logger.WithError(err).Error("Failed to read request body")
		return err
	}
	defer func(Body io.ReadCloser) {
		if closeErr := Body.Close(); closeErr != nil {
			w.logger.WithError(closeErr).Error("Failed to close request body")
		}
	}(req.Body)

	w.logger.WithFields(logrus.Fields{
		"method":  req.Method,
		"path":    req.URL.Path,
		"headers": req.Header,
		"body":    string(body),
	}).Debug("Webhook received")

	// Определяем каналы для отправки уведомлений
	channels := w.determineNotificationChannels()

	// Разбираем данные из YouTrack
	payload, parseErr := w.youtrackParser.ParseJSON(body)
	if parseErr != nil {
		w.logger.WithError(parseErr).Error("Failed to parse YouTrack data")
		return parseErr
	}

	formatter := w.youtrackParser.NewFormatter()

	// Отправляем уведомление через все выбранные каналы
	for _, channel := range channels {
		// Форматируем уведомление для конкретного канала
		formattedMessage := formatter.Format(payload, channel)

		if err = w.notificationSender.Send(channel, formattedMessage); err != nil {
			w.logger.WithError(err).WithFields(logrus.Fields{
				"channel": channel,
			}).Error("Failed to send notification to channel")
			// Продолжаем отправку в другие каналы даже при ошибке
		}
	}

	return nil
}

// determineNotificationChannels определяет каналы для отправки уведомления. Возвращает список каналов, в которые нужно отправить уведомление
func (w *WebhookService) determineNotificationChannels() []string {
	return []string{port.ChannelLogger}
}
