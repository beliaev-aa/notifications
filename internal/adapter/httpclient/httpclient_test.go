package httpclient

import (
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"net/http"
	"testing"
	"time"
)

func TestNewTelegramClient(t *testing.T) {
	type testCase struct {
		name            string
		cfg             config.TelegramConfig
		expectedTimeout time.Duration
		checkInterface  bool
	}

	testCases := []testCase{
		{
			name: "Create_TelegramClient_With_Valid_Timeout",
			cfg: config.TelegramConfig{
				BotToken: "test_token",
				Timeout:  10,
			},
			expectedTimeout: 10 * time.Second,
			checkInterface:  true,
		},
		{
			name: "Create_TelegramClient_With_Zero_Timeout",
			cfg: config.TelegramConfig{
				BotToken: "test_token",
				Timeout:  0,
			},
			expectedTimeout: 0 * time.Second,
			checkInterface:  true,
		},
		{
			name: "Create_TelegramClient_With_Large_Timeout",
			cfg: config.TelegramConfig{
				BotToken: "test_token",
				Timeout:  300,
			},
			expectedTimeout: 300 * time.Second,
			checkInterface:  true,
		},
		{
			name: "Create_TelegramClient_With_Empty_Token",
			cfg: config.TelegramConfig{
				BotToken: "",
				Timeout:  15,
			},
			expectedTimeout: 15 * time.Second,
			checkInterface:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewTelegramClient(tc.cfg)

			if client == nil {
				t.Error("expected client to be created, got: nil")
				return
			}

			if tc.checkInterface {
				if _, ok := client.(port.HTTPClient); !ok {
					t.Error("expected client to implement HTTPClient interface")
				}
			}

			httpClient, ok := client.(*http.Client)
			if !ok {
				t.Error("expected client to be *http.Client")
				return
			}

			if httpClient.Timeout != tc.expectedTimeout {
				t.Errorf("expected timeout %v, got: %v", tc.expectedTimeout, httpClient.Timeout)
			}

			if httpClient.Transport != nil {
				t.Error("expected Transport to be nil (using default), got: not nil")
			}
		})
	}
}

func TestNewVKTeamsClient(t *testing.T) {
	type testCase struct {
		name             string
		cfg              config.VKTeamsConfig
		expectedTimeout  time.Duration
		expectedInsecure bool
		checkInterface   bool
	}

	testCases := []testCase{
		{
			name: "Create_VKTeamsClient_With_InsecureSkipVerify_True",
			cfg: config.VKTeamsConfig{
				BotToken:           "test_token",
				Timeout:            30,
				ApiUrl:             "https://api.example.com/bot/v1",
				InsecureSkipVerify: true,
			},
			expectedTimeout:  30 * time.Second,
			expectedInsecure: true,
			checkInterface:   true,
		},
		{
			name: "Create_VKTeamsClient_With_InsecureSkipVerify_False",
			cfg: config.VKTeamsConfig{
				BotToken:           "test_token",
				Timeout:            20,
				ApiUrl:             "https://api.example.com/bot/v1",
				InsecureSkipVerify: false,
			},
			expectedTimeout:  20 * time.Second,
			expectedInsecure: false,
			checkInterface:   true,
		},
		{
			name: "Create_VKTeamsClient_With_Zero_Timeout",
			cfg: config.VKTeamsConfig{
				BotToken:           "test_token",
				Timeout:            0,
				ApiUrl:             "https://api.example.com/bot/v1",
				InsecureSkipVerify: false,
			},
			expectedTimeout:  0 * time.Second,
			expectedInsecure: false,
			checkInterface:   true,
		},
		{
			name: "Create_VKTeamsClient_With_Large_Timeout",
			cfg: config.VKTeamsConfig{
				BotToken:           "test_token",
				Timeout:            600,
				ApiUrl:             "https://api.example.com/bot/v1",
				InsecureSkipVerify: true,
			},
			expectedTimeout:  600 * time.Second,
			expectedInsecure: true,
			checkInterface:   true,
		},
		{
			name: "Create_VKTeamsClient_With_Empty_ApiUrl",
			cfg: config.VKTeamsConfig{
				BotToken:           "test_token",
				Timeout:            15,
				ApiUrl:             "",
				InsecureSkipVerify: false,
			},
			expectedTimeout:  15 * time.Second,
			expectedInsecure: false,
			checkInterface:   true,
		},
		{
			name: "Create_VKTeamsClient_With_Empty_Token",
			cfg: config.VKTeamsConfig{
				BotToken:           "",
				Timeout:            25,
				ApiUrl:             "https://api.example.com/bot/v1",
				InsecureSkipVerify: true,
			},
			expectedTimeout:  25 * time.Second,
			expectedInsecure: true,
			checkInterface:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewVKTeamsClient(tc.cfg)

			if client == nil {
				t.Error("expected client to be created, got: nil")
				return
			}

			if tc.checkInterface {
				if _, ok := client.(port.HTTPClient); !ok {
					t.Error("expected client to implement HTTPClient interface")
				}
			}

			httpClient, ok := client.(*http.Client)
			if !ok {
				t.Error("expected client to be *http.Client")
				return
			}

			if httpClient.Timeout != tc.expectedTimeout {
				t.Errorf("expected timeout %v, got: %v", tc.expectedTimeout, httpClient.Timeout)
			}

			if httpClient.Transport == nil {
				t.Error("expected Transport to be set, got: nil")
				return
			}

			transport, ok := httpClient.Transport.(*http.Transport)
			if !ok {
				t.Error("expected Transport to be *http.Transport")
				return
			}

			if tc.expectedInsecure {
				if transport.TLSClientConfig == nil {
					t.Error("expected TLSClientConfig to be set when InsecureSkipVerify is true, got: nil")
				} else if !transport.TLSClientConfig.InsecureSkipVerify {
					t.Error("expected InsecureSkipVerify to be true, got: false")
				}
			} else {
				if transport.TLSClientConfig != nil {
					if transport.TLSClientConfig.InsecureSkipVerify {
						t.Error("expected InsecureSkipVerify to be false, got: true")
					}
				}
			}
		})
	}
}

func TestNewTelegramClient_Interface(t *testing.T) {
	cfg := config.TelegramConfig{
		BotToken: "test_token",
		Timeout:  10,
	}

	client := NewTelegramClient(cfg)

	var _ = client

	httpClient, ok := client.(*http.Client)
	if !ok {
		t.Fatal("expected client to be *http.Client")
	}

	if httpClient == nil {
		t.Error("expected httpClient to be not nil")
	}
}

func TestNewVKTeamsClient_Interface(t *testing.T) {
	cfg := config.VKTeamsConfig{
		BotToken:           "test_token",
		Timeout:            10,
		ApiUrl:             "https://api.example.com/bot/v1",
		InsecureSkipVerify: false,
	}

	client := NewVKTeamsClient(cfg)

	var _ = client

	httpClient, ok := client.(*http.Client)
	if !ok {
		t.Fatal("expected client to be *http.Client")
	}

	if httpClient == nil {
		t.Error("expected httpClient to be not nil")
	}
}

func TestNewVKTeamsClient_TLSConfig(t *testing.T) {
	t.Run("TLS_Config_With_InsecureSkipVerify_True", func(t *testing.T) {
		cfg := config.VKTeamsConfig{
			BotToken:           "test_token",
			Timeout:            10,
			ApiUrl:             "https://api.example.com/bot/v1",
			InsecureSkipVerify: true,
		}

		client := NewVKTeamsClient(cfg)
		httpClient := client.(*http.Client)
		transport := httpClient.Transport.(*http.Transport)

		if transport.TLSClientConfig == nil {
			t.Fatal("expected TLSClientConfig to be set")
		}

		if !transport.TLSClientConfig.InsecureSkipVerify {
			t.Error("expected InsecureSkipVerify to be true")
		}

		if transport.TLSClientConfig.MinVersion != 0 {
			t.Logf("TLS MinVersion is set to %d", transport.TLSClientConfig.MinVersion)
		}
	})

	t.Run("TLS_Config_With_InsecureSkipVerify_False", func(t *testing.T) {
		cfg := config.VKTeamsConfig{
			BotToken:           "test_token",
			Timeout:            10,
			ApiUrl:             "https://api.example.com/bot/v1",
			InsecureSkipVerify: false,
		}

		client := NewVKTeamsClient(cfg)
		httpClient := client.(*http.Client)
		transport := httpClient.Transport.(*http.Transport)

		if transport.TLSClientConfig != nil {
			if transport.TLSClientConfig.InsecureSkipVerify {
				t.Error("expected InsecureSkipVerify to be false when cfg.InsecureSkipVerify is false")
			}
		}
	})
}

func TestNewTelegramClient_DefaultTransport(t *testing.T) {
	cfg := config.TelegramConfig{
		BotToken: "test_token",
		Timeout:  10,
	}

	client := NewTelegramClient(cfg)
	httpClient := client.(*http.Client)

	if httpClient.Transport != nil {
		t.Error("expected Transport to be nil (using default), got: not nil")
	}
}

func TestNewVKTeamsClient_CustomTransport(t *testing.T) {
	cfg := config.VKTeamsConfig{
		BotToken:           "test_token",
		Timeout:            10,
		ApiUrl:             "https://api.example.com/bot/v1",
		InsecureSkipVerify: true,
	}

	client := NewVKTeamsClient(cfg)
	httpClient := client.(*http.Client)

	if httpClient.Transport == nil {
		t.Fatal("expected Transport to be set, got: nil")
	}

	transport, ok := httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected Transport to be *http.Transport")
	}

	if transport == http.DefaultTransport {
		t.Error("expected custom Transport, got: http.DefaultTransport")
	}

	if transport.TLSClientConfig == nil {
		t.Error("expected TLSClientConfig to be set")
	} else {
		if !transport.TLSClientConfig.InsecureSkipVerify {
			t.Error("expected InsecureSkipVerify to be true")
		}
	}
}

func TestNewVKTeamsClient_StandardTransport(t *testing.T) {
	cfg := config.VKTeamsConfig{
		BotToken:           "test_token",
		Timeout:            10,
		ApiUrl:             "https://api.example.com/bot/v1",
		InsecureSkipVerify: false,
	}

	client := NewVKTeamsClient(cfg)
	httpClient := client.(*http.Client)

	if httpClient.Transport == nil {
		t.Fatal("expected Transport to be set, got: nil")
	}

	transport, ok := httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected Transport to be *http.Transport")
	}

	if transport == http.DefaultTransport {
		t.Error("expected custom Transport, got: http.DefaultTransport")
	}

	if transport.TLSClientConfig != nil {
		if transport.TLSClientConfig.InsecureSkipVerify {
			t.Error("expected InsecureSkipVerify to be false")
		}
	}
}
