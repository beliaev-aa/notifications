package formatter

import (
	"encoding/json"
	"fmt"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"strings"
)

const (
	nullValueString = "(Не установлен)"
)

const (
	State    string = "State"
	Priority string = "Priority"
	Assignee string = "Assignee"
	Comment  string = "Comment"
)

var fieldTranslations = map[string]string{
	State:    "Состояние",
	Priority: "Приоритет",
	Assignee: "Назначена",
	Comment:  "Комментарий",
}

// translateFieldName переводит название поля из YouTrack на русский язык
func (f *YoutrackFormatter) translateFieldName(field string) string {
	if translation, exists := fieldTranslations[field]; exists {
		return translation
	}
	return field
}

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
	return f.formatDefault(payload)
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

// formatDefault форматирует payload по умолчанию
func (f *YoutrackFormatter) formatDefault(payload *parser.YoutrackWebhookPayload) string {
	return fmt.Sprintf(`
Проект: %s
Задача: %s
Ссылка: %s
Статус: %s
Приоритет: %s
Исполнитель: %s
Автор изменения: %s
Изменения: %s`,
		f.extractFieldValue(payload.Project),
		payload.Issue.Summary,
		payload.Issue.URL,
		f.extractFieldValue(payload.Issue.State),
		f.extractFieldValue(payload.Issue.Priority),
		f.extractUserName(payload.Issue.Assignee),
		f.extractUserName(payload.Updater),
		f.extractChanges(payload.Changes))
}

func (f *YoutrackFormatter) extractFieldValue(field *parser.YoutrackFieldValue) string {
	if field == nil {
		return ""
	}
	if field.Presentation != nil && *field.Presentation != "" {
		return *field.Presentation
	}
	if field.Name != nil && *field.Name != "" {
		return *field.Name
	}
	return ""
}

// extractUserName извлекает имя пользователя
func (f *YoutrackFormatter) extractUserName(user *parser.YoutrackUser) string {
	if user == nil {
		return ""
	}

	if user.FullName != nil && *user.FullName != "" {
		return *user.FullName
	}
	if user.Login != nil && *user.Login != "" {
		return *user.Login
	}
	return ""
}

// extractChanges извлекает описание изменений
func (f *YoutrackFormatter) extractChanges(changes []parser.YoutrackChange) string {
	if len(changes) == 0 {
		return ""
	}

	var changesText []string
	for _, change := range changes {
		oldValueStr := f.extractChangeValue(change.OldValue, change.Field)
		newValueStr := f.extractChangeValue(change.NewValue, change.Field)

		if change.Field == "Comment" {
			changesText = append(changesText, fmt.Sprintf("%s: %s", f.translateFieldName(change.Field), newValueStr))
		} else {
			changesText = append(changesText, fmt.Sprintf("%s: %s → %s", f.translateFieldName(change.Field), oldValueStr, newValueStr))
		}

	}

	return strings.Join(changesText, "; ")
}

// extractChangeValue извлекает строковое значение из change value
func (f *YoutrackFormatter) extractChangeValue(value json.RawMessage, field string) string {
	if len(value) == 0 {
		return nullValueString
	}

	valueStr := strings.TrimSpace(string(value))
	if valueStr == "null" || valueStr == "" {
		return nullValueString
	}

	switch field {
	case "State", "Priority":
		var fieldValue parser.YoutrackFieldValue
		if err := json.Unmarshal(value, &fieldValue); err == nil {
			return f.extractFieldValue(&fieldValue)
		}
	case "Assignee":
		var user parser.YoutrackUser
		if err := json.Unmarshal(value, &user); err == nil {
			return f.extractUserName(&user)
		}
	case "Comment":
		var comment parser.YoutrackCommentValue
		if err := json.Unmarshal(value, &comment); err == nil {
			return f.extractCommentText(comment)
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

// extractCommentText извлекает текст комментария с упомянутыми пользователями
func (f *YoutrackFormatter) extractCommentText(comment parser.YoutrackCommentValue) string {
	text := comment.Text

	if len(comment.MentionedUsers) > 0 {
		var mentionNames []string
		for _, user := range comment.MentionedUsers {
			userName := f.extractUserName(&user)
			if userName != "" {
				mentionNames = append(mentionNames, userName)
			}
		}
		if len(mentionNames) > 0 {
			text += fmt.Sprintf(" [Упомянуты: %s]", strings.Join(mentionNames, ", "))
		}
	}

	return text
}
