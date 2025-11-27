package service

import (
	"github.com/beliaev-aa/notifications/internal/adapter/formatter"
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
func NewWebhookService(notificationSender port.NotificationSender, youtrackParser parser.YoutrackParser, logger *logrus.Logger) port.WebhookService {
	return &WebhookService{
		notificationSender: notificationSender,
		youtrackParser:     youtrackParser,
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

	// Разбираем данные из YouTrack
	payload, parseErr := w.youtrackParser.ParseJSON(body)
	if parseErr != nil {
		w.logger.WithError(parseErr).Error("Failed to parse YouTrack data")
		return parseErr
	}
	w.logger.WithFields(logrus.Fields{
		"payload": payload,
	}).Debug("Parsed payload")

	// Получаем разрешенные каналы для проекта
	channels := w.youtrackParser.GetAllowedChannels(payload)

	// Извлекаем имя проекта для логирования
	projectName := ""
	if payload.Project != nil && payload.Project.Name != nil {
		projectName = *payload.Project.Name
	}

	if len(channels) == 0 {
		if projectName != "" {
			w.logger.WithFields(logrus.Fields{
				"project": projectName,
			}).Info("Project configuration not found or no allowed channels, ignoring notification")
		} else {
			w.logger.Warn("Project name is empty in webhook payload, ignoring notification")
		}
		return nil // Игнорируем, но не возвращаем ошибку
	}

	youtrackFormatter := w.youtrackParser.NewFormatter()

	// Регистрируем специальное форматирование для Telegram канала
	youtrackFormatter.RegisterChannelFormatter(port.ChannelTelegram, formatter.FormatTelegram)
	// Регистрируем форматирование для VK Teams канала (с измененным блоком "Упомянуты:")
	youtrackFormatter.RegisterChannelFormatter(port.ChannelVKTeams, formatter.FormatVKTeams)

	// Отправляем уведомление через все выбранные каналы
	for _, channel := range channels {
		// Форматируем уведомление для конкретного канала
		formattedMessage := youtrackFormatter.Format(payload, channel)

		// Получаем chatID для каналов, которые требуют его
		chatID := ""
		if channel == port.ChannelTelegram {
			var ok bool
			chatID, ok = w.youtrackParser.GetTelegramChatID(projectName)
			if !ok || chatID == "" {
				w.logger.WithFields(logrus.Fields{
					"project": projectName,
					"channel": channel,
				}).Warn("Telegram chat ID not found for project, skipping notification")
				continue
			}
		} else if channel == port.ChannelVKTeams {
			var ok bool
			chatID, ok = w.youtrackParser.GetVKTeamsChatID(projectName)
			if !ok || chatID == "" {
				w.logger.WithFields(logrus.Fields{
					"project": projectName,
					"channel": channel,
				}).Warn("VK Teams chat ID not found for project, skipping notification")
				continue
			}
		}

		if err = w.notificationSender.Send(channel, chatID, formattedMessage); err != nil {
			w.logger.WithError(err).WithFields(logrus.Fields{
				"channel": channel,
			}).Error("Failed to send notification to channel")
			// Продолжаем отправку в другие каналы даже при ошибке
		}
	}

	return nil
}
