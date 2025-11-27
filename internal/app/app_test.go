package app

import (
	"errors"
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/tests/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestNewApp(t *testing.T) {
	type testCase struct {
		name            string
		cfg             *config.Config
		logger          *logrus.Logger
		telegramEnabled bool
		vkteamsEnabled  bool
		checkStructure  bool
	}

	validHTTPConfig := config.HTTPConfig{
		Addr:            ":8080",
		ShutdownTimeout: 10,
		ReadTimeout:     5,
		WriteTimeout:    5,
	}

	testCases := []testCase{
		{
			name: "Create_App_With_Telegram_BotToken",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
				Telegram: config.TelegramConfig{
					BotToken: "test_token",
					Timeout:  10,
				},
			},
			logger:          logrus.New(),
			telegramEnabled: true,
			vkteamsEnabled:  false,
			checkStructure:  true,
		},
		{
			name: "Create_App_Without_Telegram_BotToken",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
				Telegram: config.TelegramConfig{
					BotToken: "",
					Timeout:  10,
				},
			},
			logger:          logrus.New(),
			telegramEnabled: false,
			vkteamsEnabled:  false,
			checkStructure:  true,
		},
		{
			name: "Create_App_With_VKTeams_BotToken",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
				VKTeams: config.VKTeamsConfig{
					BotToken: "test_vkteams_token",
					Timeout:  10,
				},
			},
			logger:          logrus.New(),
			telegramEnabled: false,
			vkteamsEnabled:  true,
			checkStructure:  true,
		},
		{
			name: "Create_App_Without_VKTeams_BotToken",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
				VKTeams: config.VKTeamsConfig{
					BotToken: "",
					Timeout:  10,
					ApiUrl:   "https://api.example.com/bot/v1",
				},
			},
			logger:          logrus.New(),
			telegramEnabled: false,
			vkteamsEnabled:  false,
			checkStructure:  true,
		},
		{
			name: "Create_App_With_Both_Telegram_And_VKTeams_BotToken",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
				Telegram: config.TelegramConfig{
					BotToken: "test_telegram_token",
					Timeout:  10,
				},
				VKTeams: config.VKTeamsConfig{
					BotToken: "test_vkteams_token",
					Timeout:  10,
				},
			},
			logger:          logrus.New(),
			telegramEnabled: true,
			vkteamsEnabled:  true,
			checkStructure:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := NewApp(tc.cfg, tc.logger)

			if app == nil {
				t.Error("expected app to be created, got: nil")
				return
			}

			if tc.checkStructure {
				if app.httpServer == nil {
					t.Error("expected httpServer to be initialized, got: nil")
				}
			}
		})
	}
}

func TestApp_Run(t *testing.T) {
	type testCase struct {
		name        string
		startError  error
		expectError bool
	}

	testCases := []testCase{
		{
			name:        "App_Run_Success",
			startError:  nil,
			expectError: false,
		},
		{
			name:        "App_Run_With_Error",
			startError:  errors.New("server start error"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPServer := mocks.NewMockHTTPServer(ctrl)
			app := &App{
				httpServer: mockHTTPServer,
			}

			mockHTTPServer.EXPECT().Start().Return(tc.startError)

			err := app.Run()

			if tc.expectError {
				if err == nil {
					t.Error("expected error, got: nil")
				} else if !errors.Is(err, tc.startError) {
					t.Errorf("expected error %v, got: %v", tc.startError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestNewApp_Structure(t *testing.T) {
	type testCase struct {
		name            string
		cfg             *config.Config
		logger          *logrus.Logger
		expectedNil     bool
		checkHTTPServer bool
	}

	validHTTPConfig := config.HTTPConfig{
		Addr:            ":8080",
		ShutdownTimeout: 10,
		ReadTimeout:     5,
		WriteTimeout:    5,
	}

	testCases := []testCase{
		{
			name: "App_Structure_With_Valid_Config",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
				Telegram: config.TelegramConfig{
					BotToken: "",
					Timeout:  10,
				},
			},
			logger:          logrus.New(),
			expectedNil:     false,
			checkHTTPServer: true,
		},
		{
			name: "App_Structure_With_Telegram",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
				Telegram: config.TelegramConfig{
					BotToken: "token",
					Timeout:  5,
				},
			},
			logger:          logrus.New(),
			expectedNil:     false,
			checkHTTPServer: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := NewApp(tc.cfg, tc.logger)

			if tc.expectedNil {
				if app != nil {
					t.Errorf("expected app to be nil, got: %+v", app)
				}
			} else {
				if app == nil {
					t.Error("expected app to be created, got: nil")
					return
				}

				if tc.checkHTTPServer {
					if app.httpServer == nil {
						t.Error("expected httpServer to be initialized, got: nil")
					}
				}
			}
		})
	}
}

func TestApp_Integration(t *testing.T) {
	type testCase struct {
		name           string
		cfg            *config.Config
		logger         *logrus.Logger
		telegramConfig config.TelegramConfig
		vkteamsConfig  config.VKTeamsConfig
	}

	validHTTPConfig := config.HTTPConfig{
		Addr:            ":8080",
		ShutdownTimeout: 10,
		ReadTimeout:     5,
		WriteTimeout:    5,
	}

	testCases := []testCase{
		{
			name: "Full_Integration_Test_With_Telegram",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
			},
			logger: logrus.New(),
			telegramConfig: config.TelegramConfig{
				BotToken: "test_token",
				Timeout:  10,
			},
			vkteamsConfig: config.VKTeamsConfig{
				Timeout: 10,
				ApiUrl:  "https://api.example.com/bot/v1",
			},
		},
		{
			name: "Full_Integration_Test_Without_Telegram",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
			},
			logger: logrus.New(),
			telegramConfig: config.TelegramConfig{
				BotToken: "",
				Timeout:  10,
			},
			vkteamsConfig: config.VKTeamsConfig{
				Timeout: 10,
				ApiUrl:  "https://api.example.com/bot/v1",
			},
		},
		{
			name: "Full_Integration_Test_With_VKTeams",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
			},
			logger: logrus.New(),
			telegramConfig: config.TelegramConfig{
				Timeout: 10,
			},
			vkteamsConfig: config.VKTeamsConfig{
				BotToken: "test_vkteams_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
		},
		{
			name: "Full_Integration_Test_Without_VKTeams",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
			},
			logger: logrus.New(),
			telegramConfig: config.TelegramConfig{
				Timeout: 10,
			},
			vkteamsConfig: config.VKTeamsConfig{
				BotToken: "",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
		},
		{
			name: "Full_Integration_Test_With_Both_Telegram_And_VKTeams",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
			},
			logger: logrus.New(),
			telegramConfig: config.TelegramConfig{
				BotToken: "test_telegram_token",
				Timeout:  10,
			},
			vkteamsConfig: config.VKTeamsConfig{
				BotToken: "test_vkteams_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.cfg.Telegram = tc.telegramConfig
			tc.cfg.VKTeams = tc.vkteamsConfig

			app := NewApp(tc.cfg, tc.logger)

			if app == nil {
				t.Fatal("expected app to be created, got: nil")
			}

			if app.httpServer == nil {
				t.Fatal("expected httpServer to be initialized, got: nil")
			}
		})
	}
}
