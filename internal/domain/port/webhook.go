package port

import (
	"context"
	"net/http"
)

// WebhookService определяет порт для обработки webhook запросов
type WebhookService interface {
	ProcessWebhook(ctx context.Context, req *http.Request) error
}
