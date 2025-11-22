package usecase

import (
	"context"
	"encoding/json"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// WebhookUseCase реализует бизнес-логику обработки webhook запросов
type WebhookUseCase struct {
	logger *logrus.Logger
}

// NewWebhookUseCase создает новый экземпляр use case для обработки webhook запросов
func NewWebhookUseCase(logger *logrus.Logger) port.WebhookService {
	return &WebhookUseCase{
		logger: logger,
	}
}

// ProcessWebhook обрабатывает входящий webhook запрос: читает тело, декодирует JSON и логирует данные
func (w *WebhookUseCase) ProcessWebhook(_ context.Context, req *http.Request) error {
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

	// Декодируем JSON
	var payload map[string]interface{}
	if err = json.Unmarshal(body, &payload); err != nil {
		w.logger.WithError(err).WithField("body", string(body)).Error("Failed to decode JSON")
		return err
	}

	w.logger.WithFields(logrus.Fields{
		"method":  req.Method,
		"path":    req.URL.Path,
		"headers": req.Header,
		"payload": payload,
	}).Info("Webhook received")

	return nil
}
