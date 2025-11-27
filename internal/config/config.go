package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// filepathAbsFunc используется для тестирования - позволяет подменить проверку абсолютного пути файла
var filepathAbsFunc = filepath.Abs

const (
	// DefaultConfigPath путь к конфигурации по умолчанию
	DefaultConfigPath = "./config/config.yml"
	// EnvConfigPath переменная окружения для пути к конфигурации
	EnvConfigPath = "CONFIG_PATH"
)

// Config содержит конфигурацию приложения
type Config struct {
	HTTP          HTTPConfig          `yaml:"http"`
	Telegram      TelegramConfig      `yaml:"telegram"`
	VKTeams       VKTeamsConfig       `yaml:"vkteams"`
	Logger        LoggerConfig        `yaml:"logger"`
	Notifications NotificationsConfig `yaml:"notifications"`
}

// HTTPConfig содержит конфигурацию HTTP сервера
type HTTPConfig struct {
	Addr            string `yaml:"addr"`
	ShutdownTimeout int    `yaml:"shutdown_timeout"` // Таймаут в секундах
	ReadTimeout     int    `yaml:"read_timeout"`     // Таймаут в секундах
	WriteTimeout    int    `yaml:"write_timeout"`    // Таймаут в секундах
}

// TelegramConfig содержит глобальную конфигурацию для Telegram канала
// BotToken и Timeout используются для всех проектов
type TelegramConfig struct {
	BotToken string `yaml:"bot_token"` // Глобальный токен бота
	Timeout  int    `yaml:"timeout"`   // Таймаут для HTTP запросов к Telegram API (секунды)
}

// VKTeamsConfig содержит глобальную конфигурацию для VK Teams канала
// BotToken, Timeout и ApiUrl используются для всех проектов
type VKTeamsConfig struct {
	BotToken string `yaml:"bot_token"` // Глобальный токен бота
	Timeout  int    `yaml:"timeout"`   // Таймаут для HTTP запросов к VK Teams API (секунды)
	ApiUrl   string `yaml:"api_url"`   // URL API (обязателен)
}

// LoggerConfig содержит конфигурацию для логгера
type LoggerConfig struct {
	Level string `yaml:"level"` // Уровень логирования (debug, info, warn, error)
}

// NotificationsConfig содержит конфигурацию уведомлений
type NotificationsConfig struct {
	Youtrack YoutrackConfig `yaml:"youtrack"`
}

// YoutrackConfig содержит конфигурацию уведомлений для YouTrack
type YoutrackConfig struct {
	Projects map[string]ProjectConfig `yaml:"projects"` // Ключ - имя проекта
}

// ProjectConfig содержит конфигурацию для проекта
type ProjectConfig struct {
	// Доступные каналы для уведомлений в проекте
	AllowedChannels []string               `yaml:"allowedChannels"`
	Telegram        *ProjectTelegramConfig `yaml:"telegram,omitempty"` // Обязательно, если telegram в allowedChannels
	VKTeams         *ProjectVKTeamsConfig  `yaml:"vkteams,omitempty"`  // Обязательно, если vkteams в allowedChannels
}

// ProjectTelegramConfig настройки для Telegram
type ProjectTelegramConfig struct {
	ChatID string `yaml:"chat_id"` // Обязательное поле для каждого проекта
}

// ProjectVKTeamsConfig настройки для VK Teams
type ProjectVKTeamsConfig struct {
	ChatID string `yaml:"chat_id"` // Обязательное поле для каждого проекта
}

// LoadConfig загружает конфигурацию из YAML файла и ENV переменных
// Приоритет: ENV > YAML
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Определяем путь к конфигурации (единственное значение по умолчанию)
	configPath := os.Getenv(EnvConfigPath)
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	// Загружаем конфигурацию из YAML файла, если он существует
	if err := loadFromYAML(cfg, configPath); err != nil {
		// Если файл не найден - это не критично, продолжаем с ENV
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load config from YAML: %w", err)
		}
	}

	// Нормализуем ключи проектов к нижнему регистру
	normalizeProjectNames(cfg)

	// Перезаписываем значения из ENV переменных (приоритет ENV)
	if err := loadFromEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config from ENV: %w", err)
	}

	// Валидация конфигурации
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// loadFromYAML загружает конфигурацию из YAML файла
func loadFromYAML(cfg *Config, path string) error {
	// Преобразуем относительный путь в абсолютный
	absPath, err := filepathAbsFunc(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	data, errRead := os.ReadFile(absPath)
	if errRead != nil {
		return errRead
	}

	// Разбираем YAML напрямую в структуру Config
	if err = yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	return nil
}

// loadFromEnv загружает конфигурацию из переменных окружения
func loadFromEnv(cfg *Config) error {
	// HTTP
	// Addr
	if val := os.Getenv("HTTP_ADDR"); val != "" {
		cfg.HTTP.Addr = val
	}

	// ShutdownTimeout (значение в секундах, целое число)
	if val := os.Getenv("HTTP_SHUTDOWN_TIMEOUT"); val != "" {
		seconds, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid HTTP_SHUTDOWN_TIMEOUT format: must be integer (seconds), got: %s", val)
		}
		if seconds <= 0 {
			return fmt.Errorf("HTTP_SHUTDOWN_TIMEOUT must be positive, got: %d", seconds)
		}
		cfg.HTTP.ShutdownTimeout = seconds
	}

	// ReadTimeout (значение в секундах, целое число)
	if val := os.Getenv("HTTP_READ_TIMEOUT"); val != "" {
		seconds, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid HTTP_READ_TIMEOUT format: must be integer (seconds), got: %s", val)
		}
		if seconds <= 0 {
			return fmt.Errorf("HTTP_READ_TIMEOUT must be positive, got: %d", seconds)
		}
		cfg.HTTP.ReadTimeout = seconds
	}

	// WriteTimeout (значение в секундах, целое число)
	if val := os.Getenv("HTTP_WRITE_TIMEOUT"); val != "" {
		seconds, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid HTTP_WRITE_TIMEOUT format: must be integer (seconds), got: %s", val)
		}
		if seconds <= 0 {
			return fmt.Errorf("HTTP_WRITE_TIMEOUT must be positive, got: %d", seconds)
		}
		cfg.HTTP.WriteTimeout = seconds
	}

	// Telegram
	// BotToken
	if val := os.Getenv("TELEGRAM_BOT_TOKEN"); val != "" {
		cfg.Telegram.BotToken = val
	}

	// Timeout (значение в секундах, целое число)
	if val := os.Getenv("TELEGRAM_TIMEOUT"); val != "" {
		seconds, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid TELEGRAM_TIMEOUT format: must be integer (seconds), got: %s", val)
		}
		if seconds <= 0 {
			return fmt.Errorf("TELEGRAM_TIMEOUT must be positive, got: %d", seconds)
		}
		cfg.Telegram.Timeout = seconds
	}

	// VK Teams
	// BotToken
	if val := os.Getenv("VKTEAMS_BOT_TOKEN"); val != "" {
		cfg.VKTeams.BotToken = val
	}

	// Timeout (значение в секундах, целое число)
	if val := os.Getenv("VKTEAMS_TIMEOUT"); val != "" {
		seconds, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid VKTEAMS_TIMEOUT format: must be integer (seconds), got: %s", val)
		}
		if seconds <= 0 {
			return fmt.Errorf("VKTEAMS_TIMEOUT must be positive, got: %d", seconds)
		}
		cfg.VKTeams.Timeout = seconds
	}

	// ApiUrl
	if val := os.Getenv("VKTEAMS_API_URL"); val != "" {
		cfg.VKTeams.ApiUrl = val
	}

	// Logger.Level
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		cfg.Logger.Level = val
	}

	return nil
}

// normalizeProjectNames нормализует ключи проектов к нижнему регистру
func normalizeProjectNames(cfg *Config) {
	if cfg.Notifications.Youtrack.Projects == nil {
		// Инициализируем пустую map, если она nil
		cfg.Notifications.Youtrack.Projects = make(map[string]ProjectConfig)
		return
	}

	if len(cfg.Notifications.Youtrack.Projects) == 0 {
		// Если map пустая, ничего не делаем
		return
	}

	// Собираем оригинальные ключи для логирования
	originalKeys := make([]string, 0, len(cfg.Notifications.Youtrack.Projects))
	for k := range cfg.Notifications.Youtrack.Projects {
		originalKeys = append(originalKeys, k)
	}

	normalizedProjects := make(map[string]ProjectConfig)
	for projectName, projectConfig := range cfg.Notifications.Youtrack.Projects {
		normalizedName := strings.ToLower(projectName)
		normalizedProjects[normalizedName] = projectConfig
	}

	cfg.Notifications.Youtrack.Projects = normalizedProjects
}

// validateConfig проверяет корректность конфигурации
func validateConfig(cfg *Config) error {
	if cfg.HTTP.Addr == "" {
		return fmt.Errorf("HTTP_ADDR is required")
	}

	if cfg.HTTP.ShutdownTimeout <= 0 {
		return fmt.Errorf("HTTP_SHUTDOWN_TIMEOUT must be positive")
	}

	if cfg.HTTP.ReadTimeout <= 0 {
		return fmt.Errorf("HTTP_READ_TIMEOUT must be positive")
	}

	if cfg.HTTP.WriteTimeout <= 0 {
		return fmt.Errorf("HTTP_WRITE_TIMEOUT must be positive")
	}

	// Устанавливаем значения по умолчанию для Telegram, если не заданы
	if cfg.Telegram.Timeout <= 0 {
		cfg.Telegram.Timeout = 10
	}

	// Устанавливаем значения по умолчанию для VK Teams, если не заданы
	if cfg.VKTeams.Timeout <= 0 {
		cfg.VKTeams.Timeout = 10
	}

	// Валидация конфигурации проектов
	if err := validateNotificationsConfig(cfg); err != nil {
		return err
	}

	return nil
}

// validateNotificationsConfig проверяет корректность конфигурации уведомлений
func validateNotificationsConfig(cfg *Config) error {
	if cfg.Notifications.Youtrack.Projects == nil {
		return nil // Проекты не настроены - это допустимо (уведомления будут игнорироваться)
	}

	// Проверяем каждый проект
	for projectName, projectConfig := range cfg.Notifications.Youtrack.Projects {
		// Проверяем, что allowedChannels не пустой
		if len(projectConfig.AllowedChannels) == 0 {
			return fmt.Errorf("project %q: allowedChannels cannot be empty", projectName)
		}

		// Проверяем валидность каналов
		validChannels := map[string]bool{
			"telegram": true,
			"vkteams":  true,
			"logger":   true,
		}

		hasTelegram := false
		hasVKTeams := false
		for _, channel := range projectConfig.AllowedChannels {
			if !validChannels[channel] {
				return fmt.Errorf("project %q: invalid channel %q, allowed channels: telegram, vkteams, logger", projectName, channel)
			}
			if channel == "telegram" {
				hasTelegram = true
			}
			if channel == "vkteams" {
				hasVKTeams = true
			}
		}

		// Если telegram в allowedChannels, проверяем наличие telegram.chat_id
		if hasTelegram {
			if projectConfig.Telegram == nil {
				return fmt.Errorf("project %q: telegram.chat_id is required when telegram is in allowedChannels", projectName)
			}
			if projectConfig.Telegram.ChatID == "" {
				return fmt.Errorf("project %q: telegram.chat_id cannot be empty when telegram is in allowedChannels", projectName)
			}
			// Проверяем, что глобальный bot_token указан
			if cfg.Telegram.BotToken == "" {
				return fmt.Errorf("TELEGRAM_BOT_TOKEN is required when telegram is used in project configurations")
			}
		}

		// Если vkteams в allowedChannels, проверяем наличие vkteams.chat_id
		if hasVKTeams {
			if projectConfig.VKTeams == nil {
				return fmt.Errorf("project %q: vkteams.chat_id is required when vkteams is in allowedChannels", projectName)
			}
			if projectConfig.VKTeams.ChatID == "" {
				return fmt.Errorf("project %q: vkteams.chat_id cannot be empty when vkteams is in allowedChannels", projectName)
			}
			// Проверяем, что глобальный bot_token указан
			if cfg.VKTeams.BotToken == "" {
				return fmt.Errorf("VKTEAMS_BOT_TOKEN is required when vkteams is used in project configurations")
			}
			// Проверяем, что глобальный api_url указан
			if cfg.VKTeams.ApiUrl == "" {
				return fmt.Errorf("VKTEAMS_API_URL is required when vkteams is used in project configurations")
			}
		}
	}

	return nil
}
