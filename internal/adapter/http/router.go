package http

import (
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

// NewRouter создает новый HTTP роутер с зарегистрированными маршрутами
func NewRouter(webhookService port.WebhookService, logger *logrus.Logger) *chi.Mux {
	r := chi.NewRouter()

	h := NewHandler(webhookService, logger)

	r.Get("/health", h.Health)
	r.Post("/webhook/youtrack", h.LogPostRequest)

	return r
}
