package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// filepathAbsFunc используется для тестирования - позволяет подменить filepath.Abs
var filepathAbsFunc = filepath.Abs

const (
	// DefaultConfigPath путь к конфигурации по умолчанию
	DefaultConfigPath = "./config/config.yml"
	// EnvConfigPath переменная окружения для пути к конфигурации
	EnvConfigPath = "CONFIG_PATH"
)

// Config содержит конфигурацию приложения
type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Telegram TelegramConfig `yaml:"telegram"`
	Logger   LoggerConfig   `yaml:"logger"`
}

// HTTPConfig содержит конфигурацию HTTP сервера
type HTTPConfig struct {
	Addr            string `yaml:"addr"`
	ShutdownTimeout int    `yaml:"shutdown_timeout"` // Таймаут в секундах
	ReadTimeout     int    `yaml:"read_timeout"`     // Таймаут в секундах
	WriteTimeout    int    `yaml:"write_timeout"`    // Таймаут в секундах
}

// TelegramConfig содержит конфигурацию для Telegram канала
type TelegramConfig struct {
	Enabled  bool   `yaml:"enabled"`
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
	Timeout  int    `yaml:"timeout"` // Таймаут для HTTP запросов к Telegram API (секунды)
}

// LoggerConfig содержит конфигурацию для логгера
type LoggerConfig struct {
	Level string `yaml:"level"` // Уровень логирования (debug, info, warn, error)
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
	// Enabled
	if val := os.Getenv("TELEGRAM_ENABLED"); val != "" {
		enabled, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid TELEGRAM_ENABLED format: %w", err)
		}
		cfg.Telegram.Enabled = enabled
	}

	// BotToken
	if val := os.Getenv("TELEGRAM_BOT_TOKEN"); val != "" {
		cfg.Telegram.BotToken = val
	}

	// ChatID
	if val := os.Getenv("TELEGRAM_CHAT_ID"); val != "" {
		cfg.Telegram.ChatID = val
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

	// Logger.Level
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		cfg.Logger.Level = val
	}

	return nil
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

	// Если Telegram включен, проверяем наличие токена и chat ID
	if cfg.Telegram.Enabled {
		if cfg.Telegram.BotToken == "" {
			return fmt.Errorf("TELEGRAM_BOT_TOKEN is required when TELEGRAM_ENABLED=true")
		}
		if cfg.Telegram.ChatID == "" {
			return fmt.Errorf("TELEGRAM_CHAT_ID is required when TELEGRAM_ENABLED=true")
		}
		// Устанавливаем значения по умолчанию, если не заданы
		if cfg.Telegram.Timeout <= 0 {
			cfg.Telegram.Timeout = 10
		}
	}

	return nil
}
