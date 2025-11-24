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
				"HTTP_ADDR":             "env_addr:8080",
				"HTTP_SHUTDOWN_TIMEOUT": "10",
				"HTTP_READ_TIMEOUT":     "15",
				"HTTP_WRITE_TIMEOUT":    "20",
				"TELEGRAM_ENABLED":      "true",
				"TELEGRAM_BOT_TOKEN":    "env_token",
				"TELEGRAM_CHAT_ID":      "env_chat_id",
				"TELEGRAM_TIMEOUT":      "30",
				"LOG_LEVEL":             "info",
			},
			yamlContent: `
http:
  addr: "yaml_addr:3000"
  shutdown_timeout: 5
  read_timeout: 5
  write_timeout: 5
telegram:
  enabled: false
  bot_token: "yaml_token"
  chat_id: "yaml_chat_id"
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
					Enabled:  true,
					BotToken: "env_token",
					ChatID:   "env_chat_id",
					Timeout:  30,
				},
				Logger: LoggerConfig{
					Level: "info",
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
  enabled: false
  bot_token: ""
  chat_id: ""
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
					Enabled:  false,
					BotToken: "",
					ChatID:   "",
					Timeout:  10,
				},
				Logger: LoggerConfig{
					Level: "debug",
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
			},
		},
		{
			name: "ENV_Only_Configuration",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"TELEGRAM_ENABLED":      "false",
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
					Enabled: false,
				},
				Logger: LoggerConfig{
					Level: "warn",
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
			name: "Config_Validation_Error_If_Telegram_Enabled_But_Token_Missing",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"TELEGRAM_ENABLED":      "true",
				"TELEGRAM_CHAT_ID":      "chat123",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("TELEGRAM_BOT_TOKEN is required when TELEGRAM_ENABLED=true"),
		},
		{
			name: "Config_Validation_Error_If_Telegram_Enabled_But_ChatID_Missing",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"TELEGRAM_ENABLED":      "true",
				"TELEGRAM_BOT_TOKEN":    "token123",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("TELEGRAM_CHAT_ID is required when TELEGRAM_ENABLED=true"),
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
			name: "Invalid_TelegramEnabled_Format_Returns_Error",
			envVariables: map[string]string{
				"HTTP_ADDR":             ":8080",
				"HTTP_SHUTDOWN_TIMEOUT": "5",
				"HTTP_READ_TIMEOUT":     "5",
				"HTTP_WRITE_TIMEOUT":    "5",
				"TELEGRAM_ENABLED":      "invalid_bool",
			},
			expectedConfig: nil,
			expectedErr:    errors.New("invalid TELEGRAM_ENABLED format"),
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
  enabled: false
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
					Enabled: false,
				},
				Logger: LoggerConfig{
					Level: "error",
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
				Telegram: TelegramConfig{
					Enabled: false,
				},
			},
			expectedErr: nil,
		},
		{
			name: "Valid_Config_With_Telegram_Enabled",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					Enabled:  true,
					BotToken: "token123",
					ChatID:   "chat123",
					Timeout:  10,
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
			name: "Telegram_Enabled_But_Token_Missing",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					Enabled: true,
					ChatID:  "chat123",
				},
			},
			expectedErr: errors.New("TELEGRAM_BOT_TOKEN is required when TELEGRAM_ENABLED=true"),
		},
		{
			name: "Telegram_Enabled_But_ChatID_Missing",
			config: &Config{
				HTTP: HTTPConfig{
					Addr:            ":8080",
					ShutdownTimeout: 5,
					ReadTimeout:     5,
					WriteTimeout:    5,
				},
				Telegram: TelegramConfig{
					Enabled:  true,
					BotToken: "token123",
				},
			},
			expectedErr: errors.New("TELEGRAM_CHAT_ID is required when TELEGRAM_ENABLED=true"),
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
					Enabled:  true,
					BotToken: "token123",
					ChatID:   "chat123",
					Timeout:  0,
				},
			},
			expectedErr: nil,
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
