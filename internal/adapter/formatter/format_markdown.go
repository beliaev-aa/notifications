package formatter

import (
	"encoding/json"
	"fmt"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"strings"
)

// formatMarkdown форматирует payload для Markdown каналов с иконками и Markdown разметкой
// Принимает стратегию упоминаний и функцию для извлечения значений как параметры
func formatMarkdown(payload *parser.YoutrackWebhookPayload, mentionFormatter MentionFormatter, valueExtractor ChangeValueExtractor) string {
	var parts []string

	mention := ""
	if payload.Issue.Assignee != nil {
		mention = mentionFormatter.FormatMention(*payload.Issue.Assignee)
	}

	changed := extractChangedFromChangesWithExtractor(payload.Changes, mention, valueExtractor)
	if changed != nil {
		parts = append(parts, changed.header)
		parts = append(parts, "")
	} else {
		parts = append(parts, "")
	}

	parts = append(parts, fmt.Sprintf("*📁 Проект:* %s", escapeMarkdownV2(*payload.Project.Name)))
	parts = append(parts, fmt.Sprintf("*📋 Задача:* %s", escapeMarkdownV2(payload.Issue.Summary)))
	parts = append(parts, fmt.Sprintf("*🔗 Ссылка:* [%s](%s)", escapeMarkdownV2LinkText(payload.Issue.URL), escapeMarkdownV2URL(payload.Issue.URL)))

	if changed != nil && changed.field == State {
		parts = append(parts, fmt.Sprintf("*📊 Состояние:* %s", changed.value))
	} else {
		parts = append(parts, fmt.Sprintf("*📊 Состояние:* %s", escapeMarkdownV2(extractFieldValue(payload.Issue.State))))
	}

	if changed != nil && changed.field == Priority {
		parts = append(parts, fmt.Sprintf("*⚡️ Приоритет:* %s", changed.value))
	} else {
		parts = append(parts, fmt.Sprintf("*⚡️ Приоритет:* %s", escapeMarkdownV2(extractFieldValue(payload.Issue.Priority))))
	}

	if changed != nil && changed.field == Assignee {
		parts = append(parts, fmt.Sprintf("*👤 Назначена:* %s", changed.value))
	} else {
		parts = append(parts, fmt.Sprintf("*👤 Назначена:* %s", escapeMarkdownV2(mention)))
	}

	parts = append(parts, fmt.Sprintf("*✏️ Автор изменения:* %s", escapeMarkdownV2(extractUserName(payload.Updater))))

	if changed != nil && changed.field == Comment {
		parts = append(parts, "")
		parts = append(parts, changed.value)
	}

	return strings.Join(parts, "\n")
}

// extractChangeValueMarkdown извлекает строковое значение из change value для Markdown форматов
// Принимает функцию для обработки комментариев как параметр
func extractChangeValueMarkdown(value json.RawMessage, field string, commentExtractor CommentTextExtractor) string {
	if len(value) == 0 {
		return nullValueString
	}

	valueStr := strings.TrimSpace(string(value))
	if valueStr == "null" || valueStr == "" {
		return nullValueString
	}

	switch field {
	case State, Priority:
		var fieldValue parser.YoutrackFieldValue
		if err := json.Unmarshal(value, &fieldValue); err == nil {
			return extractFieldValue(&fieldValue)
		}
	case Assignee:
		var user parser.YoutrackUser
		if err := json.Unmarshal(value, &user); err == nil {
			return extractUserName(&user)
		}
	case Comment:
		var comment parser.YoutrackCommentValue
		if err := json.Unmarshal(value, &comment); err == nil {
			return commentExtractor(comment)
		}
	default:
		var str string
		if err := json.Unmarshal(value, &str); err == nil {
			return str
		}
		var obj map[string]interface{}
		if err := json.Unmarshal(value, &obj); err == nil {
			if name, ok := obj["name"].(string); ok && name != "" {
				return name
			}
			if val, ok := obj["value"].(string); ok && val != "" {
				return val
			}
		}
	}

	return nullValueString
}

// extractCommentTextMarkdown извлекает текст комментария с упомянутыми пользователями для Markdown форматов
// Использует MentionFormatter для форматирования упоминаний
func extractCommentTextMarkdown(comment parser.YoutrackCommentValue, formatter MentionFormatter) string {
	text := comment.Text

	for needle, replaced := range replaceSpecialCharsMap {
		text = strings.ReplaceAll(text, needle, replaced)
	}

	if len(comment.MentionedUsers) > 0 {
		var mentionNames []string
		for _, user := range comment.MentionedUsers {
			mention := formatter.FormatMention(user)
			if mention != "" {
				mentionNames = append(mentionNames, mention)
			}
		}
		if len(mentionNames) > 0 {
			text += "\n" + fmt.Sprintf("[Упомянуты: %s]", strings.Join(mentionNames, ", "))
		}
	}

	return text
}
