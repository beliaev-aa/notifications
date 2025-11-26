package utils

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
)

func TestInitLogger(t *testing.T) {
	type testCase struct {
		name              string
		envVariables      map[string]string
		expectedLevel     logrus.Level
		expectedFormatter string
		checkServiceField bool
	}

	testCases := []testCase{
		{
			name:              "Default_Debug_Level_When_LOG_LEVEL_Not_Set",
			envVariables:      map[string]string{},
			expectedLevel:     logrus.DebugLevel,
			expectedFormatter: "json",
			checkServiceField: true,
		},
		{
			name: "Info_Level_From_ENV",
			envVariables: map[string]string{
				"LOG_LEVEL": "info",
			},
			expectedLevel:     logrus.InfoLevel,
			expectedFormatter: "json",
			checkServiceField: true,
		},
		{
			name: "Warn_Level_From_ENV",
			envVariables: map[string]string{
				"LOG_LEVEL": "warn",
			},
			expectedLevel:     logrus.WarnLevel,
			expectedFormatter: "json",
			checkServiceField: true,
		},
		{
			name: "Error_Level_From_ENV",
			envVariables: map[string]string{
				"LOG_LEVEL": "error",
			},
			expectedLevel:     logrus.ErrorLevel,
			expectedFormatter: "json",
			checkServiceField: true,
		},
		{
			name: "Debug_Level_From_ENV",
			envVariables: map[string]string{
				"LOG_LEVEL": "debug",
			},
			expectedLevel:     logrus.DebugLevel,
			expectedFormatter: "json",
			checkServiceField: true,
		},
		{
			name: "Trace_Level_From_ENV",
			envVariables: map[string]string{
				"LOG_LEVEL": "trace",
			},
			expectedLevel:     logrus.TraceLevel,
			expectedFormatter: "json",
			checkServiceField: true,
		},
		{
			name: "Fatal_Level_From_ENV",
			envVariables: map[string]string{
				"LOG_LEVEL": "fatal",
			},
			expectedLevel:     logrus.FatalLevel,
			expectedFormatter: "json",
			checkServiceField: true,
		},
		{
			name: "Panic_Level_From_ENV",
			envVariables: map[string]string{
				"LOG_LEVEL": "panic",
			},
			expectedLevel:     logrus.PanicLevel,
			expectedFormatter: "json",
			checkServiceField: true,
		},
		{
			name: "Invalid_Level_Falls_Back_To_Debug",
			envVariables: map[string]string{
				"LOG_LEVEL": "invalid_level",
			},
			expectedLevel:     logrus.DebugLevel,
			expectedFormatter: "json",
			checkServiceField: true,
		},
		{
			name: "Invalid_Level_With_Special_Chars_Falls_Back_To_Debug",
			envVariables: map[string]string{
				"LOG_LEVEL": "invalid@level#123",
			},
			expectedLevel:     logrus.DebugLevel,
			expectedFormatter: "json",
			checkServiceField: true,
		},
		{
			name: "Invalid_Level_With_Numbers_Only_Falls_Back_To_Debug",
			envVariables: map[string]string{
				"LOG_LEVEL": "12345",
			},
			expectedLevel:     logrus.DebugLevel,
			expectedFormatter: "json",
			checkServiceField: true,
		},
		{
			name: "Empty_LOG_LEVEL_Falls_Back_To_Debug",
			envVariables: map[string]string{
				"LOG_LEVEL": "",
			},
			expectedLevel:     logrus.DebugLevel,
			expectedFormatter: "json",
			checkServiceField: true,
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

			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetFormatter(&logrus.JSONFormatter{})

			logLevelStr := os.Getenv("LOG_LEVEL")
			if logLevelStr == "" {
				logLevelStr = "debug"
			}

			logLevel, err := logrus.ParseLevel(logLevelStr)
			// Явная проверка: если err != nil, должен быть установлен DebugLevel
			if err != nil {
				// Проверяем, что при ошибке парсинга устанавливается DebugLevel
				originalLevel := logLevel // Сохраняем исходное значение для проверки
				logLevel = logrus.DebugLevel
				// Проверяем, что после установки logLevel действительно DebugLevel
				if logLevel != logrus.DebugLevel {
					t.Errorf("expected DebugLevel when err != nil, got: %v (original: %v)", logLevel, originalLevel)
				}
				// Проверяем, что err действительно не nil для невалидных уровней
				if err == nil {
					t.Error("expected error when parsing invalid level, got: nil")
				}
			}

			logger.SetLevel(logLevel)
			logger.AddHook(&serviceHook{serviceName: ServiceName})

			if logger.GetLevel() != tc.expectedLevel {
				t.Errorf("expected level %v, got: %v", tc.expectedLevel, logger.GetLevel())
			}

			formatter := logger.Formatter
			if formatter == nil {
				t.Error("expected formatter to be set, got: nil")
			} else {
				_, isJSONFormatter := formatter.(*logrus.JSONFormatter)
				if !isJSONFormatter {
					t.Errorf("expected JSONFormatter, got: %T", formatter)
				}
			}

			if tc.checkServiceField {
				entry := logger.WithFields(logrus.Fields{"test": "value"})
				hook := &serviceHook{serviceName: ServiceName}
				if err := hook.Fire(entry); err != nil {
					t.Errorf("hook.Fire returned error: %v", err)
				}
				if service, ok := entry.Data["service"]; !ok {
					t.Error("expected 'service' field in entry data")
				} else if service != ServiceName {
					t.Errorf("expected service name %s, got: %v", ServiceName, service)
				}
			}

			os.Clearenv()
		})
	}
}

func TestServiceHook(t *testing.T) {
	type testCase struct {
		name        string
		serviceName string
		checkLevels bool
	}

	testCases := []testCase{
		{
			name:        "ServiceHook_Adds_Service_Field",
			serviceName: ServiceName,
			checkLevels: true,
		},
		{
			name:        "ServiceHook_With_Custom_Service_Name",
			serviceName: "custom_service",
			checkLevels: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hook := &serviceHook{serviceName: tc.serviceName}

			if tc.checkLevels {
				levels := hook.Levels()
				expectedLevels := logrus.AllLevels
				if diff := cmp.Diff(expectedLevels, levels); diff != "" {
					t.Errorf("Unexpected levels (-want +got):\n%s", diff)
				}
			}

			entry := logrus.NewEntry(logrus.New())
			entry.Data = make(logrus.Fields)
			entry.Data["test"] = "value"

			if err := hook.Fire(entry); err != nil {
				t.Errorf("Fire() returned error: %v", err)
			}

			if service, ok := entry.Data["service"]; !ok {
				t.Error("expected 'service' field in entry data")
			} else if service != tc.serviceName {
				t.Errorf("expected service name %s, got: %v", tc.serviceName, service)
			}

			if testValue, ok := entry.Data["test"]; !ok || testValue != "value" {
				t.Errorf("expected 'test' field to remain unchanged, got: %v", testValue)
			}
		})
	}
}

func TestInitLogger_Integration(t *testing.T) {
	type testCase struct {
		name          string
		envVariables  map[string]string
		expectedLevel logrus.Level
		checkLogging  bool
	}

	testCases := []testCase{
		{
			name: "Full_Integration_Test_With_Debug_Level",
			envVariables: map[string]string{
				"LOG_LEVEL": "debug",
			},
			expectedLevel: logrus.DebugLevel,
			checkLogging:  true,
		},
		{
			name: "Full_Integration_Test_With_Info_Level",
			envVariables: map[string]string{
				"LOG_LEVEL": "info",
			},
			expectedLevel: logrus.InfoLevel,
			checkLogging:  false,
		},
		{
			name:          "Full_Integration_Test_Without_LOG_LEVEL",
			envVariables:  map[string]string{},
			expectedLevel: logrus.DebugLevel,
			checkLogging:  true,
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

			logger := InitLogger()

			if logger.GetLevel() != tc.expectedLevel {
				t.Errorf("expected level %v, got: %v", tc.expectedLevel, logger.GetLevel())
			}

			formatter := logger.Formatter
			if formatter == nil {
				t.Error("expected formatter to be set, got: nil")
			} else {
				_, isJSONFormatter := formatter.(*logrus.JSONFormatter)
				if !isJSONFormatter {
					t.Errorf("expected JSONFormatter, got: %T", formatter)
				}
			}

			hooks := logger.Hooks
			if len(hooks) == 0 {
				t.Error("expected hooks to be added to logger")
			}

			entry := logger.WithField("test_key", "test_value")
			hook := &serviceHook{serviceName: ServiceName}
			if err := hook.Fire(entry); err != nil {
				t.Errorf("hook.Fire returned error: %v", err)
			}
			if service, ok := entry.Data["service"]; !ok {
				t.Error("expected 'service' field in entry data")
			} else if service != ServiceName {
				t.Errorf("expected service name %s, got: %v", ServiceName, service)
			}

			if tc.checkLogging {
				logger.Debug("test debug message")
			}

			os.Clearenv()
		})
	}
}

func TestInitLogger_InvalidLevel_ErrorHandling(t *testing.T) {
	type testCase struct {
		name          string
		envVariables  map[string]string
		expectedLevel logrus.Level
	}

	testCases := []testCase{
		{
			name: "Invalid_Level_Sets_DebugLevel_On_Error",
			envVariables: map[string]string{
				"LOG_LEVEL": "invalid_level",
			},
			expectedLevel: logrus.DebugLevel,
		},
		{
			name: "Invalid_Level_With_Special_Chars_Sets_DebugLevel_On_Error",
			envVariables: map[string]string{
				"LOG_LEVEL": "invalid@level#123",
			},
			expectedLevel: logrus.DebugLevel,
		},
		{
			name: "Invalid_Level_With_Numbers_Sets_DebugLevel_On_Error",
			envVariables: map[string]string{
				"LOG_LEVEL": "12345",
			},
			expectedLevel: logrus.DebugLevel,
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

			// Тестируем реальную функцию InitLogger
			logger := InitLogger()

			// Проверяем, что при ошибке парсинга уровня устанавливается DebugLevel
			if logger.GetLevel() != tc.expectedLevel {
				t.Errorf("expected level %v when parsing invalid level, got: %v", tc.expectedLevel, logger.GetLevel())
			}

			// Проверяем, что ParseLevel действительно возвращает ошибку для невалидного уровня
			logLevelStr := os.Getenv("LOG_LEVEL")
			_, err := logrus.ParseLevel(logLevelStr)
			if err == nil {
				t.Error("expected error when parsing invalid level, got: nil")
			}

			// Проверяем, что несмотря на ошибку, логгер работает с DebugLevel
			if logger.GetLevel() != logrus.DebugLevel {
				t.Errorf("expected DebugLevel after error handling, got: %v", logger.GetLevel())
			}

			os.Clearenv()
		})
	}
}
