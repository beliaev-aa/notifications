package formatter

import (
	"encoding/json"
	"fmt"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"strings"
)

// Значение, которое будет установлено вместо null
const (
	nullValueString = "(Не установлен)"
)

// Список доступных для отслеживания полей
const (
	Assignee string = "Assignee"
	Comment  string = "Comment"
	Priority string = "Priority"
	State    string = "State"
)

// Переводы полей
var fieldTranslations = map[string]string{
	Assignee: "Назначена",
	Comment:  "Комментарий",
	Priority: "Приоритет",
	State:    "Состояние",
}

// Changed представляет информацию об изменении поля
type Changed struct {
	header string
	field  string
	value  string
}

// ChangeValueExtractor определяет тип функции для извлечения значений изменений
type ChangeValueExtractor func(json.RawMessage, string) string

// CommentTextExtractor определяет тип функции для извлечения текста комментария
type CommentTextExtractor func(parser.YoutrackCommentValue) string

// Замена экранированных спецсимволов на их не экранированные версии
var replaceSpecialCharsMap = map[string]string{
	"\\*": "*",
	"\\~": "~",
	"\\`": "`",
	"\\>": ">",
	"\\|": "|",
}

// translateFieldName переводит название поля из YouTrack на русский язык
func translateFieldName(field string) string {
	if translation, exists := fieldTranslations[field]; exists {
		return translation
	}
	return field
}

// extractFieldValue извлекает значение поля
func extractFieldValue(field *parser.YoutrackFieldValue) string {
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
func extractUserName(user *parser.YoutrackUser) string {
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

// extractChangeValue извлекает строковое значение из change value
func extractChangeValue(value json.RawMessage, field string) string {
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
			return extractCommentText(comment)
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

// extractChangedFromChangesWithExtractor извлекает информацию об изменениях из списка изменений
func extractChangedFromChangesWithExtractor(changes []parser.YoutrackChange, mention string, valueExtractor ChangeValueExtractor) *Changed {
	changed := &Changed{}

	if len(changes) == 0 {
		return nil
	}

	for _, change := range changes {
		oldValueStr := valueExtractor(change.OldValue, change.Field)
		newValueStr := valueExtractor(change.NewValue, change.Field)

		switch change.Field {
		case Assignee:
			changed.field = change.Field
			changed.header = fmt.Sprintf("*%s Изменен исполнитель задачи*", getFieldIcon(change.Field))
			changed.value = fmt.Sprintf("%s → %s", escapeMarkdownV2(oldValueStr), escapeMarkdownV2(mention))
		case Comment:
			changed.field = change.Field
			fieldIcon := getFieldIcon(change.Field)
			changed.header = fmt.Sprintf("*%s Добавлен комментарий*", fieldIcon)
			changed.value = fmt.Sprintf("*%s Комментарий*: %s", fieldIcon, escapeMarkdownV2(newValueStr))
		case Priority:
			changed.field = change.Field
			changed.header = fmt.Sprintf("*%s Изменен приоритет задачи*", getFieldIcon(change.Field))
			changed.value = fmt.Sprintf("%s → %s", escapeMarkdownV2(oldValueStr), escapeMarkdownV2(newValueStr))
		case State:
			changed.field = change.Field
			changed.header = fmt.Sprintf("*%s Изменен статус задачи*", getFieldIcon(change.Field))
			changed.value = fmt.Sprintf("%s → %s", escapeMarkdownV2(oldValueStr), escapeMarkdownV2(newValueStr))
		}
	}

	return changed
}

// extractCommentText извлекает текст комментария с упомянутыми пользователями
func extractCommentText(comment parser.YoutrackCommentValue) string {
	text := comment.Text

	for needle, replaced := range replaceSpecialCharsMap {
		text = strings.ReplaceAll(text, needle, replaced)
	}

	if len(comment.MentionedUsers) > 0 {
		var mentionNames []string
		for _, user := range comment.MentionedUsers {
			userName := extractUserName(&user)
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
