package formatter

import (
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
)

// MentionFormatter определяет интерфейс для форматирования упоминаний пользователей
type MentionFormatter interface {
	FormatMention(user parser.YoutrackUser) string
}
