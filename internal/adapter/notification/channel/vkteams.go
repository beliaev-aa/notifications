package channel

import (
	"fmt"
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// Endpoint для отправки текстовых сообщений
	vkTeamsSendTextEndpoint = "/messages/sendText"
)

// VKTeamsChannel реализует канал отправки уведомлений через VK Teams
type VKTeamsChannel struct {
	botToken string
	timeout  time.Duration
	apiURL   string // Кастомный URL API
	client   port.HTTPClient
	logger   *logrus.Logger
}

// NewVKTeamsChannel создает новый канал VK Teams
func NewVKTeamsChannel(cfg config.VKTeamsConfig, logger *logrus.Logger, httpClient port.HTTPClient) port.NotificationChannel {
	if cfg.BotToken == "" {
		logger.Warn("VK Teams bot token is empty, VK Teams channel will not work")
	}

	if cfg.ApiUrl == "" {
		logger.Error("VK Teams API URL is required, VK Teams channel will not work")
	}

	if cfg.InsecureSkipVerify {
		logger.Warn("VK Teams: SSL certificate verification is disabled. This is not recommended for production!")
	}

	timeout := time.Duration(cfg.Timeout) * time.Second

	apiURL := strings.TrimSuffix(cfg.ApiUrl, "/")

	return &VKTeamsChannel{
		botToken: cfg.BotToken,
		timeout:  timeout,
		apiURL:   apiURL,
		client:   httpClient,
		logger:   logger,
	}
}

// Send отправляет уведомление в VK Teams
func (c *VKTeamsChannel) Send(chatID string, formattedMessage string) error {
	if chatID == "" {
		return fmt.Errorf("vkteams chat ID is not configured")
	}

	// Формируем URL для отправки сообщения
	requestURL, err := url.Parse(c.apiURL + vkTeamsSendTextEndpoint)
	if err != nil {
		c.logger.WithError(err).Error("Failed to parse VK Teams API URL")
		return fmt.Errorf("failed to parse API URL: %w", err)
	}

	// Формируем параметры для GET запроса
	params := url.Values{}
	params.Set("token", c.botToken)
	params.Set("chatId", chatID)
	params.Set("text", formattedMessage)
	params.Set("parseMode", "MarkdownV2") // Используем MarkdownV2 для форматирования текста

	requestURL.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to create VK Teams GET request")
		return fmt.Errorf("failed to create GET request: %w", err)
	}

	resp, errSend := c.client.Do(req)
	if errSend != nil {
		c.logger.WithError(errSend).Error("Failed to send VK Teams message")
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
		c.logger.WithError(errRead).Warn("Failed to read VK Teams response body")
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("VK Teams API returned error")
		return fmt.Errorf("vkteams API error: status %d, response: %s", resp.StatusCode, string(body))
	}

	c.logger.WithFields(logrus.Fields{
		"chat_id": chatID,
		"status":  resp.StatusCode,
	}).Info("Notification sent via VK Teams channel")

	return nil
}

// Channel возвращает название канала
func (c *VKTeamsChannel) Channel() string {
	return port.ChannelVKTeams
}
