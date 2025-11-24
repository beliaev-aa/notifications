package channel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

const (
	// Базовый URL для Telegram Bot API
	apiUrl = "https://api.telegram.org/bot%s/%s"
	// Настройка стилизации текста сообщения
	parseMode = "MarkdownV2"
)

// TelegramChannel реализует канал отправки уведомлений через Telegram
type TelegramChannel struct {
	botToken string
	chatID   string
	timeout  time.Duration
	client   *http.Client
	logger   *logrus.Logger
}

// NewTelegramChannel создает новый канал Telegram
func NewTelegramChannel(cfg config.TelegramConfig, logger *logrus.Logger) port.NotificationChannel {
	if cfg.BotToken == "" {
		logger.Warn("Telegram bot token is empty, Telegram channel will not work")
	}
	if cfg.ChatID == "" {
		logger.Warn("Telegram chat ID is empty, Telegram channel will not work")
	}

	// Создаем HTTP клиент с таймаутом
	timeout := time.Duration(cfg.Timeout) * time.Second
	client := &http.Client{
		Timeout: timeout,
	}

	return &TelegramChannel{
		botToken: cfg.BotToken,
		chatID:   cfg.ChatID,
		timeout:  timeout,
		client:   client,
		logger:   logger,
	}
}

// Send отправляет уведомление в Telegram
func (c *TelegramChannel) Send(formattedMessage string) error {
	if c.botToken == "" {
		return fmt.Errorf("telegram bot token is not configured")
	}
	if c.chatID == "" {
		return fmt.Errorf("telegram chat ID is not configured")
	}

	// Формируем URL для отправки сообщения
	apiURL := fmt.Sprintf(apiUrl, c.botToken, "sendMessage")

	// Подготавливаем данные для отправки
	payload := map[string]interface{}{
		"chat_id":    c.chatID,
		"text":       formattedMessage,
		"parse_mode": parseMode,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal Telegram payload")
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Отправляем запрос
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.WithError(err).Error("Failed to create Telegram request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, errSend := c.client.Do(req)
	if errSend != nil {
		c.logger.WithError(errSend).Error("Failed to send Telegram message")
		return fmt.Errorf("failed to send message: %w", errSend)
	}
	defer func(Body io.ReadCloser) {
		if closeErr := Body.Close(); closeErr != nil {
			c.logger.WithError(closeErr).Error("Failed to close request body")
		}
	}(resp.Body)

	// Читаем ответ для логирования
	body, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		c.logger.WithError(errRead).Warn("Failed to read Telegram response body")
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("Telegram API returned error")
		return fmt.Errorf("telegram API error: status %d, response: %s", resp.StatusCode, string(body))
	}

	c.logger.WithFields(logrus.Fields{
		"chat_id": c.chatID,
		"status":  resp.StatusCode,
	}).Info("Notification sent via Telegram channel")

	return nil
}

// Channel возвращает название канала
func (c *TelegramChannel) Channel() string {
	return port.ChannelTelegram
}
