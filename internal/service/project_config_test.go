package service

import (
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/sirupsen/logrus"
	"strings"
	"testing"
)

func TestNewProjectConfigService(t *testing.T) {
	type testCase struct {
		name           string
		cfg            *config.Config
		logger         *logrus.Logger
		checkInterface bool
		checkNil       bool
	}

	testCases := []testCase{
		{
			name:           "Create_ProjectConfigService_Success",
			cfg:            &config.Config{},
			logger:         logrus.New(),
			checkInterface: true,
			checkNil:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewProjectConfigService(tc.cfg, tc.logger)

			if tc.checkNil {
				if service != nil {
					t.Error("expected service to be nil, got: not nil")
				}
			} else {
				if service == nil {
					t.Error("expected service to be created, got: nil")
					return
				}

				if tc.checkInterface {
					_ = service.IsProjectAllowed("test")
				}
			}
		})
	}
}

func TestProjectConfigService_GetProjectConfig(t *testing.T) {
	type testCase struct {
		name           string
		cfg            *config.Config
		projectName    string
		expectedExists bool
		expectedConfig *config.ProjectConfig
	}

	testCases := []testCase{
		{
			name: "GetProjectConfig_Project_Exists",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram", "logger"},
								Telegram: &config.ProjectTelegramConfig{
									ChatID: "123456789",
								},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedExists: true,
			expectedConfig: &config.ProjectConfig{
				AllowedChannels: []string{"telegram", "logger"},
				Telegram: &config.ProjectTelegramConfig{
					ChatID: "123456789",
				},
			},
		},
		{
			name: "GetProjectConfig_Project_Not_Exists",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName:    "project2",
			expectedExists: false,
			expectedConfig: nil,
		},
		{
			name: "GetProjectConfig_Projects_Nil",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{},
			},
			projectName:    "project1",
			expectedExists: false,
			expectedConfig: nil,
		},
		{
			name: "GetProjectConfig_Project_With_Logger_Only",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedExists: true,
			expectedConfig: &config.ProjectConfig{
				AllowedChannels: []string{"logger"},
			},
		},
		{
			name: "GetProjectConfig_Project_With_VKTeams",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams", "logger"},
								VKTeams: &config.ProjectVKTeamsConfig{
									ChatID: "chat123",
								},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedExists: true,
			expectedConfig: &config.ProjectConfig{
				AllowedChannels: []string{"vkteams", "logger"},
				VKTeams: &config.ProjectVKTeamsConfig{
					ChatID: "chat123",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewProjectConfigService(tc.cfg, logger).(*ProjectConfigServiceImpl)

			projectConfig, exists := service.GetProjectConfig(tc.projectName)

			if exists != tc.expectedExists {
				t.Errorf("expected exists %v, got: %v", tc.expectedExists, exists)
			}

			if tc.expectedConfig != nil {
				if projectConfig == nil {
					t.Error("expected project config, got: nil")
					return
				}

				if len(projectConfig.AllowedChannels) != len(tc.expectedConfig.AllowedChannels) {
					t.Errorf("expected %d channels, got: %d", len(tc.expectedConfig.AllowedChannels), len(projectConfig.AllowedChannels))
				}

				if tc.expectedConfig.Telegram != nil {
					if projectConfig.Telegram == nil {
						t.Error("expected telegram config, got: nil")
					} else if projectConfig.Telegram.ChatID != tc.expectedConfig.Telegram.ChatID {
						t.Errorf("expected chat_id %q, got: %q", tc.expectedConfig.Telegram.ChatID, projectConfig.Telegram.ChatID)
					}
				} else {
					if projectConfig.Telegram != nil {
						t.Error("expected no telegram config, got: not nil")
					}
				}
				if tc.expectedConfig.VKTeams != nil {
					if projectConfig.VKTeams == nil {
						t.Error("expected vkteams config, got: nil")
					} else if projectConfig.VKTeams.ChatID != tc.expectedConfig.VKTeams.ChatID {
						t.Errorf("expected chat_id %q, got: %q", tc.expectedConfig.VKTeams.ChatID, projectConfig.VKTeams.ChatID)
					}
				} else {
					if projectConfig.VKTeams != nil {
						t.Error("expected no vkteams config, got: not nil")
					}
				}
			} else {
				if projectConfig != nil {
					t.Errorf("expected nil config, got: %+v", projectConfig)
				}
			}
		})
	}
}

func TestProjectConfigService_IsProjectAllowed(t *testing.T) {
	type testCase struct {
		name           string
		cfg            *config.Config
		projectName    string
		expectedResult bool
	}

	testCases := []testCase{
		{
			name: "IsProjectAllowed_Project_Exists",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedResult: true,
		},
		{
			name: "IsProjectAllowed_Project_Not_Exists",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName:    "project2",
			expectedResult: false,
		},
		{
			name: "IsProjectAllowed_Projects_Nil",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{},
			},
			projectName:    "project1",
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewProjectConfigService(tc.cfg, logger)

			result := service.IsProjectAllowed(tc.projectName)

			if result != tc.expectedResult {
				t.Errorf("expected %v, got: %v", tc.expectedResult, result)
			}
		})
	}
}

func TestProjectConfigService_GetAllowedChannels(t *testing.T) {
	type testCase struct {
		name             string
		cfg              *config.Config
		projectName      string
		expectedChannels []string
	}

	testCases := []testCase{
		{
			name: "GetAllowedChannels_Project_Exists",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram", "logger"},
							},
						},
					},
				},
			},
			projectName:      "project1",
			expectedChannels: []string{"telegram", "logger"},
		},
		{
			name: "GetAllowedChannels_Project_Not_Exists",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName:      "project2",
			expectedChannels: nil,
		},
		{
			name: "GetAllowedChannels_Project_With_Logger_Only",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName:      "project1",
			expectedChannels: []string{"logger"},
		},
		{
			name: "GetAllowedChannels_Project_With_Telegram_Only",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram"},
								Telegram: &config.ProjectTelegramConfig{
									ChatID: "123456789",
								},
							},
						},
					},
				},
			},
			projectName:      "project1",
			expectedChannels: []string{"telegram"},
		},
		{
			name: "GetAllowedChannels_Project_With_VKTeams_Only",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams"},
								VKTeams: &config.ProjectVKTeamsConfig{
									ChatID: "chat123",
								},
							},
						},
					},
				},
			},
			projectName:      "project1",
			expectedChannels: []string{"vkteams"},
		},
		{
			name: "GetAllowedChannels_Project_With_VKTeams_And_Logger",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams", "logger"},
								VKTeams: &config.ProjectVKTeamsConfig{
									ChatID: "chat123",
								},
							},
						},
					},
				},
			},
			projectName:      "project1",
			expectedChannels: []string{"vkteams", "logger"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewProjectConfigService(tc.cfg, logger)

			channels := service.GetAllowedChannels(tc.projectName)

			if len(channels) != len(tc.expectedChannels) {
				t.Errorf("expected %d channels, got: %d", len(tc.expectedChannels), len(channels))
				return
			}

			for i, expected := range tc.expectedChannels {
				if i >= len(channels) {
					t.Errorf("expected channel %q at index %d, but channels length is %d", expected, i, len(channels))
					break
				}
				if channels[i] != expected {
					t.Errorf("expected channel %q at index %d, got: %q", expected, i, channels[i])
				}
			}
		})
	}
}

func TestProjectConfigService_GetTelegramChatID(t *testing.T) {
	type testCase struct {
		name           string
		cfg            *config.Config
		projectName    string
		expectedChatID string
		expectedExists bool
	}

	testCases := []testCase{
		{
			name: "GetTelegramChatID_Project_With_Telegram",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram", "logger"},
								Telegram: &config.ProjectTelegramConfig{
									ChatID: "123456789",
								},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedChatID: "123456789",
			expectedExists: true,
		},
		{
			name: "GetTelegramChatID_Project_Not_Exists",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram"},
								Telegram: &config.ProjectTelegramConfig{
									ChatID: "123456789",
								},
							},
						},
					},
				},
			},
			projectName:    "project2",
			expectedChatID: "",
			expectedExists: false,
		},
		{
			name: "GetTelegramChatID_Project_Without_Telegram_Channel",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedChatID: "",
			expectedExists: false,
		},
		{
			name: "GetTelegramChatID_Project_With_Telegram_But_No_Config",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram"},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedChatID: "",
			expectedExists: false,
		},
		{
			name: "GetTelegramChatID_Project_With_Telegram_But_Empty_ChatID",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram"},
								Telegram: &config.ProjectTelegramConfig{
									ChatID: "",
								},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedChatID: "",
			expectedExists: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)
			service := NewProjectConfigService(tc.cfg, logger)

			chatID, exists := service.GetTelegramChatID(tc.projectName)

			if exists != tc.expectedExists {
				t.Errorf("expected exists %v, got: %v", tc.expectedExists, exists)
			}

			if chatID != tc.expectedChatID {
				t.Errorf("expected chat_id %q, got: %q", tc.expectedChatID, chatID)
			}
		})
	}
}

func TestProjectConfigService_GetVKTeamsChatID(t *testing.T) {
	type testCase struct {
		name           string
		cfg            *config.Config
		projectName    string
		expectedChatID string
		expectedExists bool
	}

	testCases := []testCase{
		{
			name: "GetVKTeamsChatID_Project_With_VKTeams",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams", "logger"},
								VKTeams: &config.ProjectVKTeamsConfig{
									ChatID: "chat123",
								},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedChatID: "chat123",
			expectedExists: true,
		},
		{
			name: "GetVKTeamsChatID_Project_Not_Exists",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams"},
								VKTeams: &config.ProjectVKTeamsConfig{
									ChatID: "chat123",
								},
							},
						},
					},
				},
			},
			projectName:    "project2",
			expectedChatID: "",
			expectedExists: false,
		},
		{
			name: "GetVKTeamsChatID_Project_Without_VKTeams_Channel",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedChatID: "",
			expectedExists: false,
		},
		{
			name: "GetVKTeamsChatID_Project_With_VKTeams_But_No_Config",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams"},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedChatID: "",
			expectedExists: false,
		},
		{
			name: "GetVKTeamsChatID_Project_With_VKTeams_But_Empty_ChatID",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams"},
								VKTeams: &config.ProjectVKTeamsConfig{
									ChatID: "",
								},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedChatID: "",
			expectedExists: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel)
			service := NewProjectConfigService(tc.cfg, logger)

			chatID, exists := service.GetVKTeamsChatID(tc.projectName)

			if exists != tc.expectedExists {
				t.Errorf("expected exists %v, got: %v", tc.expectedExists, exists)
			}

			if chatID != tc.expectedChatID {
				t.Errorf("expected chat_id %q, got: %q", tc.expectedChatID, chatID)
			}
		})
	}
}

func TestProjectConfigService_Integration(t *testing.T) {
	type testCase struct {
		name        string
		cfg         *config.Config
		projectName string
	}

	testCases := []testCase{
		{
			name: "Full_Integration_Test",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram", "logger"},
								Telegram: &config.ProjectTelegramConfig{
									ChatID: "123456789",
								},
							},
							"project2": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName: "project1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewProjectConfigService(tc.cfg, logger)

			if !service.IsProjectAllowed(tc.projectName) {
				t.Errorf("expected project %q to be allowed", tc.projectName)
			}

			projectConfig, exists := service.GetProjectConfig(tc.projectName)
			if !exists {
				t.Errorf("expected project config to exist for %q", tc.projectName)
			}
			if projectConfig == nil {
				t.Fatal("expected project config, got: nil")
			}

			channels := service.GetAllowedChannels(tc.projectName)
			if len(channels) == 0 {
				t.Error("expected at least one channel")
			}

			chatID, hasChatID := service.GetTelegramChatID(tc.projectName)
			hasTelegram := false
			for _, ch := range channels {
				if ch == "telegram" {
					hasTelegram = true
					break
				}
			}

			if hasTelegram {
				if !hasChatID {
					t.Error("expected chat_id for telegram channel")
				}
				if chatID == "" {
					t.Error("expected non-empty chat_id")
				}
			}

			vkteamsChatID, hasVKTeamsChatID := service.GetVKTeamsChatID(tc.projectName)
			hasVKTeams := false
			for _, ch := range channels {
				if ch == "vkteams" {
					hasVKTeams = true
					break
				}
			}

			if hasVKTeams {
				if !hasVKTeamsChatID {
					t.Error("expected chat_id for vkteams channel")
				}
				if vkteamsChatID == "" {
					t.Error("expected non-empty vkteams chat_id")
				}
			}
		})
	}
}

func TestProjectConfigService_EdgeCases(t *testing.T) {
	type testCase struct {
		name        string
		cfg         *config.Config
		projectName string
	}

	testCases := []testCase{
		{
			name: "Empty_Project_Name",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName: "",
		},
		{
			name: "Multiple_Projects",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram", "logger"},
								Telegram: &config.ProjectTelegramConfig{
									ChatID: "111111111",
								},
							},
							"project2": {
								AllowedChannels: []string{"logger"},
							},
							"project3": {
								AllowedChannels: []string{"telegram"},
								Telegram: &config.ProjectTelegramConfig{
									ChatID: "222222222",
								},
							},
						},
					},
				},
			},
			projectName: "project2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewProjectConfigService(tc.cfg, logger)

			_ = service.IsProjectAllowed(tc.projectName)
			_, _ = service.GetProjectConfig(tc.projectName)
			_ = service.GetAllowedChannels(tc.projectName)
			_, _ = service.GetTelegramChatID(tc.projectName)
			_, _ = service.GetVKTeamsChatID(tc.projectName)
		})
	}
}

func TestProjectConfigService_CaseInsensitive(t *testing.T) {
	type testCase struct {
		name           string
		cfg            *config.Config
		projectName    string
		expectedExists bool
	}

	testCases := []testCase{
		{
			name: "GetProjectConfig_Case_Insensitive_Uppercase",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName:    "PROJECT1",
			expectedExists: true,
		},
		{
			name: "GetProjectConfig_Case_Insensitive_Mixed",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName:    "Project1",
			expectedExists: true,
		},
		{
			name: "GetProjectConfig_Case_Insensitive_Lowercase",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"PROJECT1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName:    "project1",
			expectedExists: true,
		},
		{
			name: "GetProjectConfig_Case_Insensitive_With_Telegram",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram"},
								Telegram: &config.ProjectTelegramConfig{
									ChatID: "123456789",
								},
							},
						},
					},
				},
			},
			projectName:    "PROJECT1",
			expectedExists: true,
		},
		{
			name: "GetProjectConfig_Case_Insensitive_With_VKTeams",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams"},
								VKTeams: &config.ProjectVKTeamsConfig{
									ChatID: "chat123",
								},
							},
						},
					},
				},
			},
			projectName:    "PROJECT1",
			expectedExists: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewProjectConfigService(tc.cfg, logger).(*ProjectConfigServiceImpl)

			normalizeProjectNamesInConfig(tc.cfg)

			projectConfig, exists := service.GetProjectConfig(tc.projectName)
			if exists != tc.expectedExists {
				t.Errorf("expected exists %v, got: %v", tc.expectedExists, exists)
			}
			if tc.expectedExists && projectConfig == nil {
				t.Error("expected project config, got: nil")
			}

			isAllowed := service.IsProjectAllowed(tc.projectName)
			if isAllowed != tc.expectedExists {
				t.Errorf("expected IsProjectAllowed %v, got: %v", tc.expectedExists, isAllowed)
			}

			channels := service.GetAllowedChannels(tc.projectName)
			if tc.expectedExists && len(channels) == 0 {
				t.Error("expected channels, got: empty")
			}

			if tc.expectedExists && projectConfig != nil {
				hasTelegram := false
				for _, ch := range projectConfig.AllowedChannels {
					if ch == "telegram" {
						hasTelegram = true
						break
					}
				}
				if hasTelegram {
					chatID, ok := service.GetTelegramChatID(tc.projectName)
					if !ok || chatID == "" {
						t.Error("expected telegram chat_id, got: empty")
					}
				}
				hasVKTeams := false
				for _, ch := range projectConfig.AllowedChannels {
					if ch == "vkteams" {
						hasVKTeams = true
						break
					}
				}
				if hasVKTeams {
					chatID, ok := service.GetVKTeamsChatID(tc.projectName)
					if !ok || chatID == "" {
						t.Error("expected vkteams chat_id, got: empty")
					}
				}
			}
		})
	}
}

func TestProjectConfigService_GetSendDraftNotification(t *testing.T) {
	type testCase struct {
		name              string
		cfg               *config.Config
		projectName       string
		expectedSendDraft bool
	}

	testCases := []testCase{
		{
			name: "GetSendDraftNotification_Project_Exists_With_True",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels:       []string{"logger"},
								SendDraftNotification: boolPtr(true),
							},
						},
					},
				},
			},
			projectName:       "project1",
			expectedSendDraft: true,
		},
		{
			name: "GetSendDraftNotification_Project_Exists_With_False",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels:       []string{"logger"},
								SendDraftNotification: boolPtr(false),
							},
						},
					},
				},
			},
			projectName:       "project1",
			expectedSendDraft: false,
		},
		{
			name: "GetSendDraftNotification_Project_Exists_With_Nil_Default_True",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels:       []string{"logger"},
								SendDraftNotification: nil,
							},
						},
					},
				},
			},
			projectName:       "project1",
			expectedSendDraft: true,
		},
		{
			name: "GetSendDraftNotification_Project_Not_Exists_Default_True",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			projectName:       "project2",
			expectedSendDraft: true,
		},
		{
			name: "GetSendDraftNotification_Projects_Nil_Default_True",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{},
			},
			projectName:       "project1",
			expectedSendDraft: true,
		},
		{
			name: "GetSendDraftNotification_Case_Insensitive",
			cfg: &config.Config{
				Notifications: config.NotificationsConfig{
					Youtrack: config.YoutrackConfig{
						Projects: map[string]config.ProjectConfig{
							"project1": {
								AllowedChannels:       []string{"logger"},
								SendDraftNotification: boolPtr(false),
							},
						},
					},
				},
			},
			projectName:       "PROJECT1",
			expectedSendDraft: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewProjectConfigService(tc.cfg, logger)

			result := service.GetSendDraftNotification(tc.projectName)

			if result != tc.expectedSendDraft {
				t.Errorf("expected sendDraftNotification %v, got: %v", tc.expectedSendDraft, result)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func normalizeProjectNamesInConfig(cfg *config.Config) {
	if cfg.Notifications.Youtrack.Projects == nil {
		return
	}

	normalizedProjects := make(map[string]config.ProjectConfig)
	for projectName, projectConfig := range cfg.Notifications.Youtrack.Projects {
		normalizedName := strings.ToLower(projectName)
		normalizedProjects[normalizedName] = projectConfig
	}

	cfg.Notifications.Youtrack.Projects = normalizedProjects
}
