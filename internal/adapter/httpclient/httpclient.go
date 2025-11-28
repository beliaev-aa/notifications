package httpclient

import (
	"crypto/tls"
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"net/http"
	"time"
)

// NewTelegramClient создает HTTP клиент для Telegram канала
func NewTelegramClient(cfg config.TelegramConfig) port.HTTPClient {
	return &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}
}

// NewVKTeamsClient создает HTTP клиент для VK Teams канала с настройками TLS
func NewVKTeamsClient(cfg config.VKTeamsConfig) port.HTTPClient {
	var transport *http.Transport

	// Настраиваем TLS для игнорирования проверки сертификата, если указано
	if cfg.InsecureSkipVerify {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	} else {
		// Используем стандартный Transport с проверкой сертификата
		transport = &http.Transport{}
	}

	return &http.Client{
		Timeout:   time.Duration(cfg.Timeout) * time.Second,
		Transport: transport,
	}
}
