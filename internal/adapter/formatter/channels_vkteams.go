package formatter

import (
	"encoding/json"
	"fmt"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
)

// VKTeamsMentionFormatter форматирует упоминания для VK Teams
type VKTeamsMentionFormatter struct{}

// FormatMention форматирует упоминание пользователя для VK Teams
func (f *VKTeamsMentionFormatter) FormatMention(user parser.YoutrackUser) string {
	if user.Email != nil && *user.Email != "" {
		return fmt.Sprintf("@[%s]", *user.Email)
	} else if user.Login != nil && *user.Login != "" {
		return fmt.Sprintf("@%s", *user.Login)
	}
	return ""
}

// FormatVKTeams форматирует payload для VK Teams канала с иконками и Markdown разметкой
func FormatVKTeams(payload *parser.YoutrackWebhookPayload) string {
	return formatMarkdown(payload, &VKTeamsMentionFormatter{}, extractChangeValueVKTeams)
}

// extractChangeValueVKTeams извлекает строковое значение из change value для VK Teams
func extractChangeValueVKTeams(value json.RawMessage, field string) string {
	return extractChangeValueMarkdown(value, field, extractCommentTextVKTeams)
}

// extractCommentTextVKTeams извлекает текст комментария с упомянутыми пользователями для VK Teams
func extractCommentTextVKTeams(comment parser.YoutrackCommentValue) string {
	return extractCommentTextMarkdown(comment, &VKTeamsMentionFormatter{})
}
