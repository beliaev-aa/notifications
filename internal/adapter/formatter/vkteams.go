package formatter

import (
	"encoding/json"
	"fmt"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"strings"
)

// FormatVKTeams форматирует payload для VK Teams канала с иконками и Markdown разметкой
func FormatVKTeams(payload *parser.YoutrackWebhookPayload) string {
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

	if changes := extractChangesVKTeams(payload.Changes); changes != "" {
		parts = append(parts, "")
		parts = append(parts, fmt.Sprintf("🔄 *Изменения:*\n%s", changes))
	}

	return strings.Join(parts, "\n")
}

// extractChangesVKTeams извлекает описание изменений для VK Teams с иконками
func extractChangesVKTeams(changes []parser.YoutrackChange) string {
	if len(changes) == 0 {
		return ""
	}

	var changesText []string
	for _, change := range changes {
		oldValueStr := extractChangeValueVKTeams(change.OldValue, change.Field)
		newValueStr := extractChangeValueVKTeams(change.NewValue, change.Field)

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

// extractChangeValueVKTeams извлекает строковое значение из change value для VK Teams
func extractChangeValueVKTeams(value json.RawMessage, field string) string {
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
			return extractCommentTextVKTeams(comment)
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

// extractCommentTextVKTeams извлекает текст комментария с упомянутыми пользователями для VK Teams
func extractCommentTextVKTeams(comment parser.YoutrackCommentValue) string {
	text := comment.Text

	for needle, replaced := range replaceSpecialCharsMap {
		text = strings.ReplaceAll(text, needle, replaced)
	}

	if len(comment.MentionedUsers) > 0 {
		var mentionNames []string
		for _, user := range comment.MentionedUsers {
			var mention string
			if user.Email != nil && *user.Email != "" {
				mention = fmt.Sprintf("%s", *user.Email)
			} else if user.Login != nil && *user.Login != "" {
				mention = fmt.Sprintf("@%s", *user.Login)
			}
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
