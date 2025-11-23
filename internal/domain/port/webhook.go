package port

import (
	"net/http"
)

// WebhookService определяет порт для обработки webhook запросов
type WebhookService interface {
	ProcessWebhook(req *http.Request) error
}
