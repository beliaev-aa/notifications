package formatter

import (
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
)

// YoutrackFormatter реализует порт YoutrackFormatter для форматирования YouTrack уведомлений
type YoutrackFormatter struct {
	channelFormatters map[string]func(payload *parser.YoutrackWebhookPayload) string
}

// NewYoutrackFormatter создает новый экземпляр для YouTrack
func NewYoutrackFormatter() parser.YoutrackFormatter {
	return &YoutrackFormatter{
		channelFormatters: make(map[string]func(payload *parser.YoutrackWebhookPayload) string),
	}
}

// Format форматирует payload для указанного канала
func (f *YoutrackFormatter) Format(payload *parser.YoutrackWebhookPayload, channel string) string {
	// Если есть специфичное форматирование для канала - используем его
	if formatter, exists := f.channelFormatters[channel]; exists {
		return formatter(payload)
	}

	// Иначе используем форматирование по умолчанию
	return formatDefault(payload)
}

// RegisterChannelFormatter регистрирует специфичное форматирование для канала
func (f *YoutrackFormatter) RegisterChannelFormatter(channel string, formatter func(payload *parser.YoutrackWebhookPayload) string) {
	if channel == "" {
		return
	}
	if formatter == nil {
		return
	}
	f.channelFormatters[channel] = formatter
}

