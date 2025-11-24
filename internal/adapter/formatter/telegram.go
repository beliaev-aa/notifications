package formatter

import (
	"fmt"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"strings"
)

// escapeMarkdownV2 экранируем текст по спецификации MarkdownV2 для Telegram
func escapeMarkdownV2(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// escapeMarkdownV2 экранируем текст ссылки по спецификации MarkdownV2 для Telegram
func escapeMarkdownV2LinkText(text string) string {
	specialChars := []string{"_", "*", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// escapeMarkdownV2 экранируем текст ссылки по спецификации MarkdownV2 для Telegram
func escapeMarkdownV2URL(url string) string {
	result := url
	result = strings.ReplaceAll(result, "\\", "\\\\")
	result = strings.ReplaceAll(result, ")", "\\)")
	return result
}

// FormatTelegram форматирует payload для Telegram канала с иконками и Markdown разметкой
func FormatTelegram(payload *parser.YoutrackWebhookPayload) string {
	var parts []string

	for _, change := range payload.Changes {
		switch change.Field {
		case Assignee:
			parts = append(parts, "*👤 Изменен исполнитель задачи*")
		case Comment:
			parts = append(parts, "*💬 Добавлен комментарий*")
		case Priority:
			parts = append(parts, "*⚡️ Изменен приоритет задачи*")
		case State:
			parts = append(parts, "*📊 Изменен статус задачи*")
		}
	}

	parts = append(parts, "")

	parts = append(parts, fmt.Sprintf("*📁 Проект:* %s", escapeMarkdownV2(*payload.Project.Name)))
	parts = append(parts, fmt.Sprintf("*📋 Задача:* %s", escapeMarkdownV2(payload.Issue.Summary)))
	parts = append(parts, fmt.Sprintf("🔗 *Ссылка:* [%s](%s)", escapeMarkdownV2LinkText(payload.Issue.URL), escapeMarkdownV2URL(payload.Issue.URL)))
	parts = append(parts, fmt.Sprintf("*📊 Состояние:* %s", escapeMarkdownV2(extractFieldValue(payload.Issue.State))))
	parts = append(parts, fmt.Sprintf("*⚡️ Приоритет:* %s", escapeMarkdownV2(extractFieldValue(payload.Issue.Priority))))
	parts = append(parts, fmt.Sprintf("*👤 Назначена:* %s", escapeMarkdownV2(extractUserName(payload.Issue.Assignee))))
	parts = append(parts, fmt.Sprintf("*✏️ Автор изменения:* %s", escapeMarkdownV2(extractUserName(payload.Updater))))

	if changes := extractChangesTelegram(payload.Changes); changes != "" {
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("🔄 *Изменения:*\n%s", changes))
	}

	return strings.Join(parts, "\n")
}

// extractChangesTelegram извлекает описание изменений для Telegram с иконками в виде ненумерованного списка
func extractChangesTelegram(changes []parser.YoutrackChange) string {
	if len(changes) == 0 {
		return ""
	}

	var changesText []string
	for _, change := range changes {
		oldValueStr := extractChangeValue(change.OldValue, change.Field)
		newValueStr := extractChangeValue(change.NewValue, change.Field)

		fieldIcon := getFieldIcon(change.Field)
		fieldName := translateFieldName(change.Field)

		if change.Field == Comment {
			changesText = append(changesText, fmt.Sprintf("%s *%s:* %s", fieldIcon, fieldName, escapeMarkdownV2(newValueStr)))
		} else {
			changesText = append(changesText, fmt.Sprintf("%s *%s:* %s → %s", fieldIcon, fieldName, escapeMarkdownV2(oldValueStr), escapeMarkdownV2(newValueStr)))
		}
	}

	return strings.Join(changesText, "\n")
}

// getFieldIcon возвращает иконку для поля
func getFieldIcon(field string) string {
	icons := map[string]string{
		State:    "📊",
		Priority: "⚡",
		Assignee: "👤",
		Comment:  "💬",
	}
	if icon, exists := icons[field]; exists {
		return icon
	}
	return "📝"
}
