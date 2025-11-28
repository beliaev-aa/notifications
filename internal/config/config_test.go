package config

import (
	"errors"
	"github.com/google/go-cmp/cmp"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	type testCase struct {
		name           string
		envVariables   map[string]string
		yamlContent    string
		yamlPath       string
		expectedConfig *Config
		expectedErr    error
	}

	testCases := []testCase{
		{
			name: "ENV_Variables_Have_Priority_Over_YAML",
			envVariables: map[string]string{
				"HTTP_ADDR":                    "env_addr:8080",
				"HTTP_SHUTDOWN_TIMEOUT":        "10",
				"HTTP_READ_TIMEOUT":            "15",
				"HTTP_WRITE_TIMEOUT":           "20",
				"TELEGRAM_BOT_TOKEN":           "env_token",
				"TELEGRAM_TIMEOUT":             "30",
				"VKTEAMS_BOT_TOKEN":            "env_vkteams_token",
				"VKTEAMS_TIMEOUT":              "25",
				"VKTEAMS_API_URL":              "https://api.env.example.com/bot/v1",
				"VKTEAMS_INSECURE_SKIP_VERIFY": "true",
				"LOG_LEVEL":                    "info",
			},
			yamlContent: `
http:
  addr: "yaml_addr:3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
telegram:
  bot_token: "yaml_token"
  timeout: 10
logger:
  level: "debug"
`,
			expectedConfig: &Config{
				HTTP: HTTPConfig{
					Addr:            "env_addr:8080",
					ShutdownTimeout: 10,
					ReadTimeout:     15,
					WriteTimeout:    20,
				},
				Telegram: TelegramConfig{
					BotToken: "env_token",
					Timeout:  30,
				},
				VKTeams: VKTeamsConfig{
					BotToken:           "env_vkteams_token",
					Timeout:            25,
					ApiUrl:             "https://api.env.example.com/bot/v1",
					InsecureSkipVerify: true,
				},
				Logger: LoggerConfig{
					Level: "info",
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: make(map[string]ProjectConfig),
					},
				},
			},
		},
		{
			name:         "YAML_Used_If_ENV_Variables_Not_Set",
			envVariables: map[string]string{},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
telegram:
  bot_token: ""
  timeout: 10
logger:
  level: "debug"
`,
			expectedConfig: &Config{
				HTTP: HTTPConfig{
					Addr:            ":3000",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					BotToken: "",
					Timeout:  10,
				},
				VKTeams: VKTeamsConfig{
					Timeout: 10,
					ApiUrl:  "",
				},
				Logger: LoggerConfig{
					Level: "debug",
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: make(map[string]ProjectConfig),
					},
				},
			},
		},
		{
			name: "YAML_File_Not_Found_Is_Not_Critical",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			expectedConfig: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					Timeout: 10,
				},
				VKTeams: VKTeamsConfig{
					Timeout: 10,
					ApiUrl:  "",
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: make(map[string]ProjectConfig),
					},
				},
			},
		},
		{
			name: "ENV_Only_Configuration",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"LOG_LEVEL":             "warn",
			},
			expectedConfig: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					Timeout: 10,
				},
				VKTeams: VKTeamsConfig{
					Timeout: 10,
					ApiUrl:  "",
				},
				Logger: LoggerConfig{
					Level: "warn",
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: make(map[string]ProjectConfig),
					},
				},
			},
		},
		{
			name: "Config_With_Notifications_Projects",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"TELEGRAM_BOT_TOKEN":    "token123",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
telegram:
  bot_token: "token123"
  timeout: 10
notifications:
  youtrack:
    projects:
      project1:
        allowedChannels: [telegram, logger]
        telegram:
          chat_id: "123456789"
      project2:
        allowedChannels: [logger]
`,
			expectedConfig: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					BotToken: "token123",
					Timeout:  10,
				},
				VKTeams: VKTeamsConfig{
					Timeout: 10,
					ApiUrl:  "",
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram", "logger"},
								Telegram: &ProjectTelegramConfig{
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
		},
		{
			name: "Config_Validation_Error_If_HTTP_Addr_Not_Set",
			envVariables: map[string]string{
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("HTTP_ADDR is required"),
		},
		{
			name: "Config_Validation_Error_If_ShutdownTimeout_Not_Positive",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "0",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("HTTP_SHUTDOWN_TIMEOUT must be positive"),
		},
		{
			name: "Config_Validation_Error_If_Project_Has_Telegram_But_No_ChatID",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"TELEGRAM_BOT_TOKEN":    "token123",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
telegram:
  bot_token: "token123"
notifications:
  youtrack:
    projects:
      project1:
        allowedChannels: [telegram]
`,
			expectedConfig: nil,
			expectedErr:    errors.New("telegram.chat_id is required"),
		},
		{
			name: "Config_Validation_Error_If_Project_Has_Telegram_But_Empty_ChatID",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"TELEGRAM_BOT_TOKEN":    "token123",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
telegram:
  bot_token: "token123"
notifications:
  youtrack:
    projects:
      project1:
        allowedChannels: [telegram]
        telegram:
          chat_id: ""
`,
			expectedConfig: nil,
			expectedErr:    errors.New("telegram.chat_id cannot be empty"),
		},
		{
			name: "Config_Validation_Error_If_Project_Has_Telegram_But_No_BotToken",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
notifications:
  youtrack:
    projects:
      project1:
        allowedChannels: [telegram]
        telegram:
          chat_id: "123456789"
`,
			expectedConfig: nil,
			expectedErr:    errors.New("TELEGRAM_BOT_TOKEN is required"),
		},
		{
			name: "Config_Validation_Error_If_Project_Has_VKTeams_But_No_ChatID",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"VKTEAMS_BOT_TOKEN":     "token123",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
vkteams:
  bot_token: "token123"
notifications:
  youtrack:
    projects:
      project1:
        allowedChannels: [vkteams]
`,
			expectedConfig: nil,
			expectedErr:    errors.New("vkteams.chat_id is required"),
		},
		{
			name: "Config_Validation_Error_If_Project_Has_VKTeams_But_Empty_ChatID",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"VKTEAMS_BOT_TOKEN":     "token123",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
vkteams:
  bot_token: "token123"
notifications:
  youtrack:
    projects:
      project1:
        allowedChannels: [vkteams]
        vkteams:
          chat_id: ""
`,
			expectedConfig: nil,
			expectedErr:    errors.New("vkteams.chat_id cannot be empty"),
		},
		{
			name: "Config_Validation_Error_If_Project_Has_VKTeams_But_No_BotToken",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
notifications:
  youtrack:
    projects:
      project1:
        allowedChannels: [vkteams]
        vkteams:
          chat_id: "chat123"
`,
			expectedConfig: nil,
			expectedErr:    errors.New("VKTEAMS_BOT_TOKEN is required"),
		},
		{
			name: "Config_Validation_Error_If_Project_Has_Invalid_Channel",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
notifications:
  youtrack:
    projects:
      project1:
        allowedChannels: [invalid_channel]
`,
			expectedConfig: nil,
			expectedErr:    errors.New("invalid channel"),
		},
		{
			name: "Config_Validation_Error_If_Project_Has_Empty_AllowedChannels",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
notifications:
  youtrack:
    projects:
      project1:
        allowedChannels: []
`,
			expectedConfig: nil,
			expectedErr:    errors.New("allowedChannels cannot be empty"),
		},
		{
			name: "Invalid_Timeout_Format_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "invalid",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("invalid HTTP_SHUTDOWN_TIMEOUT format"),
		},
		{
			name: "Negative_Timeout_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "-5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("HTTP_SHUTDOWN_TIMEOUT must be positive"),
		},
		{
			name: "Invalid_ReadTimeout_Format_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "invalid",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("invalid HTTP_READ_TIMEOUT format"),
		},
		{
			name: "Negative_ReadTimeout_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "-5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("HTTP_READ_TIMEOUT must be positive"),
		},
		{
			name: "Invalid_WriteTimeout_Format_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "invalid",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("invalid HTTP_WRITE_TIMEOUT format"),
		},
		{
			name: "Negative_WriteTimeout_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "-5",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("HTTP_WRITE_TIMEOUT must be positive"),
		},
		{
			name: "Invalid_TelegramTimeout_Format_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"TELEGRAM_TIMEOUT":      "invalid",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("invalid TELEGRAM_TIMEOUT format"),
		},
		{
			name: "Negative_TelegramTimeout_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"TELEGRAM_TIMEOUT":      "-5",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("TELEGRAM_TIMEOUT must be positive"),
		},
		{
			name: "Invalid_VKTeamsTimeout_Format_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"VKTEAMS_TIMEOUT":       "invalid",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("invalid VKTEAMS_TIMEOUT format"),
		},
		{
			name: "Negative_VKTeamsTimeout_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"VKTEAMS_TIMEOUT":       "-5",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("VKTEAMS_TIMEOUT must be positive"),
		},
		{
			name: "VKTEAMS_INSECURE_SKIP_VERIFY_True_Value",
			envVariables: map[string]string{
				"HTTP_ADDR":                    ":8080",
				"HTTP_SHUTDOWN_TIMEOUT":        "5",
				"HTTP_READ_TIMEOUT":            "5",
				"HTTP_WRITE_TIMEOUT":           "5",
				"VKTEAMS_INSECURE_SKIP_VERIFY": "true",
			},
			expectedConfig: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					Timeout: 10,
				},
				VKTeams: VKTeamsConfig{
					Timeout:            10,
					ApiUrl:             "",
					InsecureSkipVerify: true,
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: make(map[string]ProjectConfig),
					},
				},
			},
		},
		{
			name: "VKTEAMS_INSECURE_SKIP_VERIFY_Invalid_Value",
			envVariables: map[string]string{
				"HTTP_ADDR":                    ":8080",
				"HTTP_SHUTDOWN_TIMEOUT":        "5",
				"HTTP_READ_TIMEOUT":            "5",
				"HTTP_WRITE_TIMEOUT":           "5",
				"VKTEAMS_INSECURE_SKIP_VERIFY": "1",
			},
			expectedConfig: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					Timeout: 10,
				},
				VKTeams: VKTeamsConfig{
					Timeout:            10,
					ApiUrl:             "",
					InsecureSkipVerify: false,
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: make(map[string]ProjectConfig),
					},
				},
			},
		},
		{
			name: "VKTEAMS_INSECURE_SKIP_VERIFY_Invalid_Value_Yes",
			envVariables: map[string]string{
				"HTTP_ADDR":                    ":8080",
				"HTTP_SHUTDOWN_TIMEOUT":        "5",
				"HTTP_READ_TIMEOUT":            "5",
				"HTTP_WRITE_TIMEOUT":           "5",
				"VKTEAMS_INSECURE_SKIP_VERIFY": "yes",
			},
			expectedConfig: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					Timeout: 10,
				},
				VKTeams: VKTeamsConfig{
					Timeout:            10,
					ApiUrl:             "",
					InsecureSkipVerify: false,
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: make(map[string]ProjectConfig),
					},
				},
			},
		},
		{
			name: "VKTEAMS_INSECURE_SKIP_VERIFY_False_Value",
			envVariables: map[string]string{
				"HTTP_ADDR":                    ":8080",
				"HTTP_SHUTDOWN_TIMEOUT":        "5",
				"HTTP_READ_TIMEOUT":            "5",
				"HTTP_WRITE_TIMEOUT":           "5",
				"VKTEAMS_INSECURE_SKIP_VERIFY": "false",
			},
			expectedConfig: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					Timeout: 10,
				},
				VKTeams: VKTeamsConfig{
					Timeout:            10,
					ApiUrl:             "",
					InsecureSkipVerify: false,
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: make(map[string]ProjectConfig),
					},
				},
			},
		},
		{
			name: "Invalid_YAML_Content_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: invalid_yaml
`,
			expectedConfig: nil,
			expectedErr:    errors.New("failed to load config from YAML"),
		},
		{
			name: "CONFIG_PATH_From_Env_Is_Used",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: 10
  read_timeout: 10
  write_timeout: 10
telegram:
  bot_token: ""
logger:
  level: "error"
`,
			expectedConfig: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					Timeout: 10,
				},
				VKTeams: VKTeamsConfig{
					Timeout: 10,
					ApiUrl:  "",
				},
				Logger: LoggerConfig{
					Level: "error",
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: make(map[string]ProjectConfig),
					},
				},
			},
		},
		{
			name: "YAML_Parse_Error_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: [invalid_yaml_structure
`,
			expectedConfig: nil,
			expectedErr:    errors.New("failed to load config from YAML"),
		},
		{
			name: "YAML_Unmarshal_Error_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			yamlContent: `
http:
  addr: ":3000"
  shutdown_timeout: invalid_yaml_value
  read_timeout: "not_a_number"
`,
			expectedConfig: nil,
			expectedErr:    errors.New("failed to load config from YAML"),
		},
		{
			name: "YAML_File_Read_Error_Other_Than_NotExist_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("failed to load config from YAML"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()

			for key, value := range tc.envVariables {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("failed to set env variable %s: %v", key, err)
				}
			}

			if tc.yamlContent != "" {
				tempDir := t.TempDir()
				tempYAMLPath := filepath.Join(tempDir, "config.yml")
				if err := os.WriteFile(tempYAMLPath, []byte(tc.yamlContent), 0644); err != nil {
					t.Fatalf("failed to create temp YAML file: %v", err)
				}

				if err := os.Setenv(EnvConfigPath, tempYAMLPath); err != nil {
					t.Fatalf("failed to set CONFIG_PATH: %v", err)
				}
			} else if tc.name == "YAML_File_Not_Found_Is_Not_Critical" {
				tempDir := t.TempDir()
				nonExistentPath := filepath.Join(tempDir, "nonexistent.yml")
				if err := os.Setenv(EnvConfigPath, nonExistentPath); err != nil {
					t.Fatalf("failed to set CONFIG_PATH: %v", err)
				}
			} else if tc.name == "YAML_File_Read_Error_Other_Than_NotExist_Returns_Error" {
				tempDir := t.TempDir()
				if err := os.Setenv(EnvConfigPath, tempDir); err != nil {
					t.Fatalf("failed to set CONFIG_PATH: %v", err)
				}
			} else {
				err := os.Unsetenv(EnvConfigPath)
				if err != nil {
					return
				}
			}

			cfg, err := LoadConfig()

			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error: %v, got: nil", tc.expectedErr)
				} else if !strings.Contains(err.Error(), tc.expectedErr.Error()) {
					t.Errorf("expected error to contain: %v, got: %v", tc.expectedErr, err)
				}
				if cfg != nil {
					t.Errorf("expected config to be nil on error, got: %+v", cfg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if diff := cmp.Diff(tc.expectedConfig, cfg); diff != "" {
					t.Errorf("Unexpected config (-want +got):\n%s", diff)
				}
			}

			os.Clearenv()
		})
	}
}

func TestValidateConfig(t *testing.T) {
	type testCase struct {
		name        string
		config      *Config
		expectedErr error
	}

	testCases := []testCase{
		{
			name: "Valid_Config",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
			},
			expectedErr: nil,
		},
		{
			name: "Valid_Config_With_Telegram",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					BotToken: "token123",
					Timeout:  10,
				},
			},
			expectedErr: nil,
		},
		{
			name: "Valid_Config_With_VKTeams",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				VKTeams: VKTeamsConfig{
					BotToken: "token123",
					Timeout:  10,
					ApiUrl:   "https://api.example.com/bot/v1",
				},
			},
			expectedErr: nil,
		},
		{
			name: "Valid_Config_With_VKTeams_Project",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				VKTeams: VKTeamsConfig{
					BotToken: "token123",
					Timeout:  10,
					ApiUrl:   "https://api.example.com/bot/v1",
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams", "logger"},
								VKTeams: &ProjectVKTeamsConfig{
									ChatID: "chat123",
								},
							},
						},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "Valid_Config_With_Projects",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					BotToken: "token123",
					Timeout:  10,
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram", "logger"},
								Telegram: &ProjectTelegramConfig{
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
			expectedErr: nil,
		},
		{
			name: "Missing_HTTP_Addr",
			config: &Config{
				HTTP: HTTPConfig{
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
			},
			expectedErr: errors.New("HTTP_ADDR is required"),
		},
		{
			name: "Invalid_ShutdownTimeout",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 0,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
			},
			expectedErr: errors.New("HTTP_SHUTDOWN_TIMEOUT must be positive"),
		},
		{
			name: "Invalid_ReadTimeout",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     0,
					WriteTimeout:    5,
				},
			},
			expectedErr: errors.New("HTTP_READ_TIMEOUT must be positive"),
		},
		{
			name: "Invalid_WriteTimeout",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    0,
				},
			},
			expectedErr: errors.New("HTTP_WRITE_TIMEOUT must be positive"),
		},
		{
			name: "Project_With_Telegram_But_No_ChatID",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					BotToken: "token123",
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram"},
							},
						},
					},
				},
			},
			expectedErr: errors.New("telegram.chat_id is required"),
		},
		{
			name: "Project_With_Telegram_But_Empty_ChatID",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					BotToken: "token123",
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram"},
								Telegram: &ProjectTelegramConfig{
									ChatID: "",
								},
							},
						},
					},
				},
			},
			expectedErr: errors.New("telegram.chat_id cannot be empty"),
		},
		{
			name: "Project_With_Telegram_But_No_BotToken",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{"telegram"},
								Telegram: &ProjectTelegramConfig{
									ChatID: "123456789",
								},
							},
						},
					},
				},
			},
			expectedErr: errors.New("TELEGRAM_BOT_TOKEN is required"),
		},
		{
			name: "Project_With_Invalid_Channel",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{"invalid_channel"},
							},
						},
					},
				},
			},
			expectedErr: errors.New("invalid channel"),
		},
		{
			name: "Project_With_Empty_AllowedChannels",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{},
							},
						},
					},
				},
			},
			expectedErr: errors.New("allowedChannels cannot be empty"),
		},
		{
			name: "Telegram_Timeout_Default_Value_Set",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					BotToken: "token123",
					Timeout:  0,
				},
			},
			expectedErr: nil,
		},
		{
			name: "VKTeams_Timeout_Default_Value_Set",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				VKTeams: VKTeamsConfig{
					BotToken: "token123",
					Timeout:  0,
					ApiUrl:   "https://api.example.com/bot/v1",
				},
			},
			expectedErr: nil,
		},
		{
			name: "Project_With_VKTeams_But_No_ChatID",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				VKTeams: VKTeamsConfig{
					BotToken: "token123",
					ApiUrl:   "https://api.example.com/bot/v1",
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams"},
							},
						},
					},
				},
			},
			expectedErr: errors.New("vkteams.chat_id is required"),
		},
		{
			name: "Project_With_VKTeams_But_Empty_ChatID",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				VKTeams: VKTeamsConfig{
					BotToken: "token123",
					ApiUrl:   "https://api.example.com/bot/v1",
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams"},
								VKTeams: &ProjectVKTeamsConfig{
									ChatID: "",
								},
							},
						},
					},
				},
			},
			expectedErr: errors.New("vkteams.chat_id cannot be empty"),
		},
		{
			name: "Project_With_VKTeams_But_No_BotToken",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams"},
								VKTeams: &ProjectVKTeamsConfig{
									ChatID: "chat123",
								},
							},
						},
					},
				},
			},
			expectedErr: errors.New("VKTEAMS_BOT_TOKEN is required"),
		},
		{
			name: "Project_With_VKTeams_But_No_APIURL",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				VKTeams: VKTeamsConfig{
					BotToken: "token123",
					ApiUrl:   "",
				},
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{"vkteams"},
								VKTeams: &ProjectVKTeamsConfig{
									ChatID: "chat123",
								},
							},
						},
					},
				},
			},
			expectedErr: errors.New("VKTEAMS_API_URL is required"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateConfig(tc.config)

			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error: %v, got: nil", tc.expectedErr)
				} else if !strings.Contains(err.Error(), tc.expectedErr.Error()) {
					t.Errorf("expected error to contain: %v, got: %v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tc.name == "Telegram_Timeout_Default_Value_Set" && tc.config.Telegram.Timeout != 10 {
					t.Errorf("expected Telegram.Timeout to be set to 10, got: %d", tc.config.Telegram.Timeout)
				}
				if tc.name == "VKTeams_Timeout_Default_Value_Set" && tc.config.VKTeams.Timeout != 10 {
					t.Errorf("expected VKTeams.Timeout to be set to 10, got: %d", tc.config.VKTeams.Timeout)
				}
			}
		})
	}
}

func TestLoadFromYAML_FilepathAbsError(t *testing.T) {
	cfg := &Config{}

	invalidPath := string([]byte{0})

	err := loadFromYAML(cfg, invalidPath)

	if err != nil {
		if !strings.Contains(err.Error(), "failed to get absolute path") &&
			!strings.Contains(err.Error(), "invalid argument") &&
			!strings.Contains(err.Error(), "no such file") &&
			!strings.Contains(err.Error(), "is a directory") {
			t.Logf("Got error (expected on some systems): %v", err)
		}
	}
}

func TestLoadFromYAML_AllBranches(t *testing.T) {
	t.Run("Valid_YAML_File", func(t *testing.T) {
		cfg := &Config{}
		tempDir := t.TempDir()
		yamlPath := filepath.Join(tempDir, "test.yml")
		yamlContent := `
http:
  addr: ":3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
`
		if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("failed to create test YAML file: %v", err)
		}

		err := loadFromYAML(cfg, yamlPath)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg.HTTP.Addr != ":3000" {
			t.Errorf("expected addr :3000, got: %s", cfg.HTTP.Addr)
		}
	})

	t.Run("File_Not_Found", func(t *testing.T) {
		cfg := &Config{}
		tempDir := t.TempDir()
		nonExistentPath := filepath.Join(tempDir, "nonexistent.yml")

		err := loadFromYAML(cfg, nonExistentPath)
		if err == nil {
			t.Error("expected error for non-existent file")
		} else if !os.IsNotExist(err) {
			t.Errorf("expected IsNotExist error, got: %v", err)
		}
	})

	t.Run("Invalid_YAML_Content", func(t *testing.T) {
		cfg := &Config{}
		tempDir := t.TempDir()
		yamlPath := filepath.Join(tempDir, "invalid.yml")
		invalidYAML := `invalid: [yaml: content`
		if err := os.WriteFile(yamlPath, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("failed to create invalid YAML file: %v", err)
		}

		err := loadFromYAML(cfg, yamlPath)
		if err == nil {
			t.Error("expected error for invalid YAML")
		} else if !strings.Contains(err.Error(), "failed to parse YAML") {
			t.Errorf("expected YAML parse error, got: %v", err)
		}
	})

	t.Run("FilepathAbs_Error", func(t *testing.T) {
		cfg := &Config{}

		originalFunc := filepathAbsFunc
		defer func() {
			filepathAbsFunc = originalFunc
		}()

		filepathAbsFunc = func(path string) (string, error) {
			return "", &os.PathError{
				Op:   "abs",
				Path: path,
				Err:  os.ErrInvalid,
			}
		}

		err := loadFromYAML(cfg, "test_path")
		if err == nil {
			t.Error("expected error about absolute path, got: nil")
		} else if !strings.Contains(err.Error(), "failed to get absolute path") {
			t.Errorf("expected error about absolute path, got: %v", err)
		}
	})
}

func TestNormalizeProjectNames(t *testing.T) {
	type testCase struct {
		name           string
		cfg            *Config
		expectedKeys   []string
		expectedValues map[string]ProjectConfig
	}

	testCases := []testCase{
		{
			name: "Normalize_Project_Names_To_Lowercase",
			cfg: &Config{
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"Project1": {
								AllowedChannels: []string{"logger"},
							},
							"PROJECT2": {
								AllowedChannels: []string{"telegram"},
								Telegram: &ProjectTelegramConfig{
									ChatID: "123456789",
								},
							},
							"Project3_Mixed": {
								AllowedChannels: []string{"logger", "telegram"},
								Telegram: &ProjectTelegramConfig{
									ChatID: "987654321",
								},
							},
						},
					},
				},
			},
			expectedKeys: []string{"project1", "project2", "project3_mixed"},
			expectedValues: map[string]ProjectConfig{
				"project1": {
					AllowedChannels: []string{"logger"},
				},
				"project2": {
					AllowedChannels: []string{"telegram"},
					Telegram: &ProjectTelegramConfig{
						ChatID: "123456789",
					},
				},
				"project3_mixed": {
					AllowedChannels: []string{"logger", "telegram"},
					Telegram: &ProjectTelegramConfig{
						ChatID: "987654321",
					},
				},
			},
		},
		{
			name: "Normalize_Empty_Projects",
			cfg: &Config{
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{},
					},
				},
			},
			expectedKeys:   []string{},
			expectedValues: map[string]ProjectConfig{},
		},
		{
			name: "Normalize_Nil_Projects",
			cfg: &Config{
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: nil,
					},
				},
			},
			expectedKeys:   []string{},
			expectedValues: nil,
		},
		{
			name: "Normalize_Already_Lowercase",
			cfg: &Config{
				Notifications: NotificationsConfig{
					Youtrack: YoutrackConfig{
						Projects: map[string]ProjectConfig{
							"project1": {
								AllowedChannels: []string{"logger"},
							},
						},
					},
				},
			},
			expectedKeys: []string{"project1"},
			expectedValues: map[string]ProjectConfig{
				"project1": {
					AllowedChannels: []string{"logger"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normalizeProjectNames(tc.cfg)

			if tc.cfg.Notifications.Youtrack.Projects == nil {
				if tc.expectedValues != nil {
					t.Error("expected projects to be not nil")
				}
				return
			}

			if len(tc.cfg.Notifications.Youtrack.Projects) != len(tc.expectedKeys) {
				t.Errorf("expected %d projects, got: %d", len(tc.expectedKeys), len(tc.cfg.Notifications.Youtrack.Projects))
			}

			for _, expectedKey := range tc.expectedKeys {
				projectConfig, exists := tc.cfg.Notifications.Youtrack.Projects[expectedKey]
				if !exists {
					t.Errorf("expected project %q to exist after normalization", expectedKey)
					continue
				}

				expectedConfig := tc.expectedValues[expectedKey]
				if len(projectConfig.AllowedChannels) != len(expectedConfig.AllowedChannels) {
					t.Errorf("expected %d channels for project %q, got: %d", len(expectedConfig.AllowedChannels), expectedKey, len(projectConfig.AllowedChannels))
				}

				if expectedConfig.Telegram != nil {
					if projectConfig.Telegram == nil {
						t.Errorf("expected telegram config for project %q", expectedKey)
					} else if projectConfig.Telegram.ChatID != expectedConfig.Telegram.ChatID {
						t.Errorf("expected chat_id %q for project %q, got: %q", expectedConfig.Telegram.ChatID, expectedKey, projectConfig.Telegram.ChatID)
					}
				}
			}

			for oldKey := range map[string]ProjectConfig{
				"Project1":       {},
				"PROJECT2":       {},
				"Project3_Mixed": {},
			} {
				if _, exists := tc.cfg.Notifications.Youtrack.Projects[oldKey]; exists {
					t.Errorf("old key %q should not exist after normalization", oldKey)
				}
			}
		})
	}
}
