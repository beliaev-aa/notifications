package port

import "github.com/beliaev-aa/notifications/internal/config"

// ProjectConfigService определяет порт для работы с конфигурацией проекта
type ProjectConfigService interface {
	// GetProjectConfig получение конфигурации проекта
	GetProjectConfig(projectName string) (*config.ProjectConfig, bool)
	// IsProjectAllowed проверка, разрешен ли проект
	IsProjectAllowed(projectName string) bool
	// GetAllowedChannels получение разрешенных каналов
	GetAllowedChannels(projectName string) []string
	// GetTelegramChatID получение chat_id для проекта
	GetTelegramChatID(projectName string) (string, bool)
	// GetVKTeamsChatID получение chat_id для VK Teams канала проекта
	GetVKTeamsChatID(projectName string) (string, bool)
}
