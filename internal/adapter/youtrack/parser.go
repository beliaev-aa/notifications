package youtrack

import (
	"encoding/json"
	"fmt"
	"github.com/beliaev-aa/notifications/internal/adapter/formatter"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"strings"
)

// Parser парсит данные YouTrack webhook
type Parser struct {
	projectConfigService port.ProjectConfigService
}

// NewParser создает новый парсер
func NewParser(projectConfigService port.ProjectConfigService) parser.YoutrackParser {
	return &Parser{
		projectConfigService: projectConfigService,
	}
}

// ParseJSON парсит JSON данные YouTrack webhook и возвращает структурированный payload
func (p *Parser) ParseJSON(payload []byte) (*parser.YoutrackWebhookPayload, error) {
	var webhookPayload parser.YoutrackWebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook payload: %w", err)
	}

	return &webhookPayload, nil
}

// NewFormatter создает форматтер для YouTrack payload
func (p *Parser) NewFormatter() parser.YoutrackFormatter {
	return formatter.NewYoutrackFormatter()
}

// GetAllowedChannels возвращает разрешенные каналы для проекта из payload
func (p *Parser) GetAllowedChannels(payload *parser.YoutrackWebhookPayload) []string {
	if payload.Project == nil || payload.Project.Name == nil || *payload.Project.Name == "" {
		return []string{}
	}

	// Нормализуем имя проекта к нижнему регистру
	projectName := strings.ToLower(*payload.Project.Name)
	if !p.projectConfigService.IsProjectAllowed(projectName) {
		return []string{}
	}

	return p.projectConfigService.GetAllowedChannels(projectName)
}

// GetTelegramChatID возвращает chat_id для Telegram канала проекта
func (p *Parser) GetTelegramChatID(projectName string) (string, bool) {
	return p.projectConfigService.GetTelegramChatID(strings.ToLower(projectName))
}

// GetVKTeamsChatID возвращает chat_id для VK Teams канала проекта
func (p *Parser) GetVKTeamsChatID(projectName string) (string, bool) {
	return p.projectConfigService.GetVKTeamsChatID(strings.ToLower(projectName))
}

// GetSendDraftNotification возвращает настройку отправки уведомлений для черновиков проекта
func (p *Parser) GetSendDraftNotification(projectName string) bool {
	return p.projectConfigService.GetSendDraftNotification(strings.ToLower(projectName))
}
