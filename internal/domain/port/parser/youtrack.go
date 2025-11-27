package parser

import "encoding/json"

// YoutrackWebhookPayload payload от YouTrack webhook
type YoutrackWebhookPayload struct {
	Project *YoutrackFieldValue `json:"project"`
	Issue   YoutrackIssue       `json:"issue"`
	Updater *YoutrackUser       `json:"updater"`
	Changes []YoutrackChange    `json:"changes"`
}

// YoutrackUser представляет пользователя YouTrack
type YoutrackUser struct {
	FullName *string `json:"fullName"`
	Login    *string `json:"login"`
	Email    *string `json:"email"`
}

// YoutrackFieldValue представляет значение поля (State, Priority)
type YoutrackFieldValue struct {
	Name         *string `json:"name"`
	Presentation *string `json:"presentation"`
}

// YoutrackCommentValue представляет значение комментария
type YoutrackCommentValue struct {
	Text           string         `json:"text"`
	MentionedUsers []YoutrackUser `json:"mentionedUsers"`
}

// YoutrackChange представляет одно изменение в задаче
type YoutrackChange struct {
	Field    string          `json:"field"`
	OldValue json.RawMessage `json:"oldValue"`
	NewValue json.RawMessage `json:"newValue"`
}

// YoutrackIssue представляет задачу YouTrack
type YoutrackIssue struct {
	Summary  string              `json:"summary"`
	URL      string              `json:"url"`
	State    *YoutrackFieldValue `json:"state"`
	Priority *YoutrackFieldValue `json:"priority"`
	Assignee *YoutrackUser       `json:"assignee"`
}

// YoutrackParser определяет порт для парсинга данных YouTrack webhook
type YoutrackParser interface {
	// ParseJSON парсит JSON данные YouTrack webhook и возвращает структурированный payload
	ParseJSON(payload []byte) (*YoutrackWebhookPayload, error)
	// NewFormatter создает форматтер для данного типа payload
	NewFormatter() YoutrackFormatter
	// GetAllowedChannels возвращает разрешенные каналы для проекта из payload
	// Если проект не разрешен или имя проекта пустое, возвращает пустой слайс
	GetAllowedChannels(payload *YoutrackWebhookPayload) []string
	// GetTelegramChatID возвращает chat_id для Telegram канала проекта
	// Возвращает chat_id и true, если проект разрешен и имеет Telegram конфигурацию, иначе пустую строку и false
	GetTelegramChatID(projectName string) (string, bool)
	// GetVKTeamsChatID возвращает chat_id для VK Teams канала проекта
	// Возвращает chat_id и true, если проект разрешен и имеет VK Teams конфигурацию, иначе пустую строку и false
	GetVKTeamsChatID(projectName string) (string, bool)
}

// YoutrackFormatter определяет порт для форматирования YouTrack payload для различных каналов
type YoutrackFormatter interface {
	// Format форматирует YouTrack payload для указанного канала
	// Если для канала есть специфичное форматирование - использует его, иначе форматирование по умолчанию
	Format(payload *YoutrackWebhookPayload, channel string) string
	// RegisterChannelFormatter регистрирует специфичное форматирование для канала
	RegisterChannelFormatter(channel string, formatter func(payload *YoutrackWebhookPayload) string)
}
