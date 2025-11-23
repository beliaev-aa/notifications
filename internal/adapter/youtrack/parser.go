package youtrack

import (
	"encoding/json"
	"fmt"
	"github.com/beliaev-aa/notifications/internal/adapter/formatter"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
)

// Parser парсит данные YouTrack webhook
type Parser struct{}

// NewParser создает новый парсер
func NewParser() parser.YoutrackParser {
	return &Parser{}
}

// ParseJSON парсит JSON данные YouTrack webhook и возвращает структурированный payload
func (p *Parser) ParseJSON(payload []byte) (*parser.YoutrackWebhookPayload, error) {
	var webhookPayload parser.YoutrackWebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook payload: %w", err)
	}

	return &webhookPayload, nil
}

// NewFormatter создает форматтер для YouTrack payload
func (p *Parser) NewFormatter() parser.YoutrackFormatter {
	return formatter.NewYoutrackFormatter()
}
