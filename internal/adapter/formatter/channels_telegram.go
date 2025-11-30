package formatter

import (
	"encoding/json"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
)

// TelegramMentionFormatter форматирует упоминания для Telegram
type TelegramMentionFormatter struct{}

// FormatMention форматирует упоминание пользователя для Telegram
func (f *TelegramMentionFormatter) FormatMention(user parser.YoutrackUser) string {
	if user.FullName != nil && *user.FullName != "" {
		return *user.FullName
	} else if user.Email != nil && *user.Email != "" {
		return *user.Email
	} else if user.Login != nil && *user.Login != "" {
		return *user.Login
	}
	return ""
}

// FormatTelegram форматирует payload для Telegram канала с иконками и Markdown разметкой
func FormatTelegram(payload *parser.YoutrackWebhookPayload) string {
	return formatMarkdown(payload, &TelegramMentionFormatter{}, extractChangeValueTelegram)
}

// extractChangeValueTelegram извлекает строковое значение из change value для Telegram
func extractChangeValueTelegram(value json.RawMessage, field string) string {
	return extractChangeValueMarkdown(value, field, extractCommentTextTelegram)
}

// extractCommentTextTelegram извлекает текст комментария с упомянутыми пользователями для Telegram
func extractCommentTextTelegram(comment parser.YoutrackCommentValue) string {
	return extractCommentTextMarkdown(comment, &TelegramMentionFormatter{})
}
