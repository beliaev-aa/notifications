package app

import (
	"bytes"
	"errors"
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/tests/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"strings"
	"testing"
)

func TestNewApp(t *testing.T) {
	type testCase struct {
		name            string
		cfg             *config.Config
		logger          *logrus.Logger
		telegramEnabled bool
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
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.cfg.Telegram = tc.telegramConfig

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

func TestNewApp_ProjectConfigurations_Logging(t *testing.T) {
	type testCase struct {
		name               string
		cfg                *config.Config
		logger             *logrus.Logger
		expectedLogMessage string
		expectedProjects   []string
		expectedCount      int
		checkProjectsInLog bool
	}

	validHTTPConfig := config.HTTPConfig{
		Addr:            ":8080",
		ShutdownTimeout: 10,
		ReadTimeout:     5,
		WriteTimeout:    5,
	}

	testCases := []testCase{
		{
			name: "Log_Loaded_Project_Configurations_With_Projects",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
				Telegram: config.TelegramConfig{
					BotToken: "",
					Timeout:  10,
				},
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"demo": {
								AllowedChannels: []string{"logger"},
							},
							"project2": {
								AllowedChannels: []string{"telegram"},
								Telegram: &config.ProjectTelegramConfig{
									ChatID: "123456789",
								},
							},
						},
					},
				},
			},
			logger:             logrus.New(),
			expectedLogMessage: "Loaded project configurations",
			expectedProjects:   []string{"demo", "project2"},
			expectedCount:      2,
			checkProjectsInLog: true,
		},
		{
			name: "Log_No_Project_Configurations_When_Projects_Nil",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
				Telegram: config.TelegramConfig{
					BotToken: "",
					Timeout:  10,
				},
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: nil,
					},
				},
			},
			logger:             logrus.New(),
			expectedLogMessage: "No project configurations found in config file",
			expectedProjects:   nil,
			expectedCount:      0,
			checkProjectsInLog: false,
		},
		{
			name: "Log_No_Project_Configurations_When_Projects_Empty",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
				Telegram: config.TelegramConfig{
					BotToken: "",
					Timeout:  10,
				},
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: make(map[string]config.ProjectConfig),
					},
				},
			},
			logger:             logrus.New(),
			expectedLogMessage: "No project configurations found in config file",
			expectedProjects:   nil,
			expectedCount:      0,
			checkProjectsInLog: false,
		},
		{
			name: "Log_Loaded_Project_Configurations_Single_Project",
			cfg: &config.Config{
				HTTP: validHTTPConfig,
				Telegram: config.TelegramConfig{
					BotToken: "",
					Timeout:  10,
				},
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"demo": {
								AllowedChannels: []string{"logger", "telegram"},
								Telegram: &config.ProjectTelegramConfig{
									ChatID: "123456789",
								},
							},
						},
					},
				},
			},
			logger:             logrus.New(),
			expectedLogMessage: "Loaded project configurations",
			expectedProjects:   []string{"demo"},
			expectedCount:      1,
			checkProjectsInLog: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetFormatter(&logrus.JSONFormatter{})
			logger.SetLevel(logrus.InfoLevel)

			app := NewApp(tc.cfg, logger)

			if app == nil {
				t.Error("expected app to be created, got: nil")
				return
			}

			logOutput := buf.String()

			if !strings.Contains(logOutput, tc.expectedLogMessage) {
				t.Errorf("expected log message %q, got output: %s", tc.expectedLogMessage, logOutput)
			}

			if tc.checkProjectsInLog {
				for _, project := range tc.expectedProjects {
					if !strings.Contains(logOutput, project) {
						t.Errorf("expected project %q in log output, got: %s", project, logOutput)
					}
				}

				if !strings.Contains(logOutput, `"count"`) {
					t.Error("expected 'count' field in log output")
				}

				if !strings.Contains(logOutput, `"projects"`) {
					t.Error("expected 'projects' field in log output")
				}
			}
		})
	}
}
