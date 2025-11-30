package formatter

import (
	"fmt"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"strings"
)

// formatDefault форматирует payload по умолчанию
func formatDefault(payload *parser.YoutrackWebhookPayload) string {
	return fmt.Sprintf(`
Проект: %s
Задача: %s
Ссылка: %s
Статус: %s
Приоритет: %s
Исполнитель: %s
Автор изменения: %s
Изменения: %s`,
		extractFieldValue(payload.Project),
		payload.Issue.Summary,
		payload.Issue.URL,
		extractFieldValue(payload.Issue.State),
		extractFieldValue(payload.Issue.Priority),
		extractUserName(payload.Issue.Assignee),
		extractUserName(payload.Updater),
		extractChangesDefault(payload.Changes))
}

// extractChangesDefault извлекает описание изменений для форматирования по умолчанию
func extractChangesDefault(changes []parser.YoutrackChange) string {
	if len(changes) == 0 {
		return ""
	}

	var changesText []string
	for _, change := range changes {
		oldValueStr := extractChangeValue(change.OldValue, change.Field)
		newValueStr := extractChangeValue(change.NewValue, change.Field)

		if change.Field == Comment {
			changesText = append(changesText, fmt.Sprintf("%s: %s", translateFieldName(change.Field), newValueStr))
		} else {
			changesText = append(changesText, fmt.Sprintf("%s: %s → %s", translateFieldName(change.Field), oldValueStr, newValueStr))
		}
	}

	return strings.Join(changesText, "; ")
}
