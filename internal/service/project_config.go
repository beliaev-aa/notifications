package service

import (
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/sirupsen/logrus"
	"strings"
)

// ProjectConfigServiceImpl реализует ProjectConfigService для работы с конфигурацией проектов
type ProjectConfigServiceImpl struct {
	cfg    *config.Config
	logger *logrus.Logger
}

// NewProjectConfigService создает новый экземпляр сервиса конфигурации проектов
func NewProjectConfigService(cfg *config.Config, logger *logrus.Logger) port.ProjectConfigService {
	return &ProjectConfigServiceImpl{
		cfg:    cfg,
		logger: logger,
	}
}

// GetProjectConfig получает конфигурацию проекта по имени
func (s *ProjectConfigServiceImpl) GetProjectConfig(projectName string) (*config.ProjectConfig, bool) {
	if s.cfg.Notifications.Youtrack.Projects == nil {
		s.logger.WithFields(logrus.Fields{
			"project":            projectName,
			"project_normalized": strings.ToLower(projectName),
		}).Debug("Projects map is nil")
		return nil, false
	}

	normalizedName := strings.ToLower(projectName)
	projectConfig, exists := s.cfg.Notifications.Youtrack.Projects[normalizedName]
	if !exists {
		availableProjects := make([]string, 0, len(s.cfg.Notifications.Youtrack.Projects))
		for p := range s.cfg.Notifications.Youtrack.Projects {
			availableProjects = append(availableProjects, p)
		}
		s.logger.WithFields(logrus.Fields{
			"project":            projectName,
			"project_normalized": normalizedName,
			"available_projects": availableProjects,
		}).Debug("Project not found in configuration")
		return nil, false
	}

	return &projectConfig, true
}

// IsProjectAllowed проверяет, разрешен ли проект (есть ли конфигурация для него)
func (s *ProjectConfigServiceImpl) IsProjectAllowed(projectName string) bool {
	_, exists := s.GetProjectConfig(projectName)
	return exists
}

// GetAllowedChannels получает список разрешенных каналов для проекта
func (s *ProjectConfigServiceImpl) GetAllowedChannels(projectName string) []string {
	projectConfig, exists := s.GetProjectConfig(projectName)
	if !exists {
		return nil
	}

	return projectConfig.AllowedChannels
}

// GetTelegramChatID получает chat_id для Telegram канала проекта
func (s *ProjectConfigServiceImpl) GetTelegramChatID(projectName string) (string, bool) {
	projectConfig, exists := s.GetProjectConfig(projectName)
	if !exists {
		return "", false
	}

	hasTelegram := false
	for _, channel := range projectConfig.AllowedChannels {
		if channel == "telegram" {
			hasTelegram = true
			break
		}
	}

	if !hasTelegram {
		return "", false
	}

	if projectConfig.Telegram == nil {
		s.logger.WithFields(logrus.Fields{
			"project": projectName,
		}).Warn("Telegram channel is in allowedChannels but telegram config is missing")
		return "", false
	}

	if projectConfig.Telegram.ChatID == "" {
		s.logger.WithFields(logrus.Fields{
			"project": projectName,
		}).Warn("Telegram channel is in allowedChannels but chat_id is empty")
		return "", false
	}

	return projectConfig.Telegram.ChatID, true
}

// GetVKTeamsChatID получает chat_id для VK Teams канала проекта
func (s *ProjectConfigServiceImpl) GetVKTeamsChatID(projectName string) (string, bool) {
	projectConfig, exists := s.GetProjectConfig(projectName)
	if !exists {
		return "", false
	}

	hasVKTeams := false
	for _, channel := range projectConfig.AllowedChannels {
		if channel == "vkteams" {
			hasVKTeams = true
			break
		}
	}

	if !hasVKTeams {
		return "", false
	}

	if projectConfig.VKTeams == nil {
		s.logger.WithFields(logrus.Fields{
			"project": projectName,
		}).Warn("VK Teams channel is in allowedChannels but vkteams config is missing")
		return "", false
	}

	if projectConfig.VKTeams.ChatID == "" {
		s.logger.WithFields(logrus.Fields{
			"project": projectName,
		}).Warn("VK Teams channel is in allowedChannels but chat_id is empty")
		return "", false
	}

	return projectConfig.VKTeams.ChatID, true
}
