package http

import (
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Handler обрабатывает HTTP запросы
type Handler struct {
	webhookService port.WebhookService
	logger         *logrus.Logger
}

// NewHandler создает новый HTTP handler
func NewHandler(webhookService port.WebhookService, logger *logrus.Logger) *Handler {
	return &Handler{
		webhookService: webhookService,
		logger:         logger,
	}
}

// Health обрабатывает запрос проверки состояния сервиса
func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte("ok"))
	if err != nil {
		h.logger.WithError(err).Error("Failed to write health response")
		http.Error(w, "Failed to write health response", http.StatusInternalServerError)
		return
	}
}

// YoutrackWebhook обрабатывает POST запросы из webhook YouTrack
func (h *Handler) YoutrackWebhook(w http.ResponseWriter, r *http.Request) {
	// Делегируем обработку бизнес-логики
	if err := h.webhookService.ProcessWebhook(r); err != nil {
		h.logger.WithError(err).Error("Failed to process webhook")
		http.Error(w, "Failed to process webhook", http.StatusBadRequest)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		h.logger.WithError(err).Error("Failed to write response")
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
