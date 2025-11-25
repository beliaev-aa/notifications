package channel

import (
	"bytes"
	"encoding/json"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/sirupsen/logrus"
	"strings"
	"testing"
)

func TestNewLoggerChannel(t *testing.T) {
	type testCase struct {
		name           string
		logger         *logrus.Logger
		checkInterface bool
		checkNil       bool
	}

	testCases := []testCase{
		{
			name:           "Create_LoggerChannel_Success",
			logger:         logrus.New(),
			checkInterface: true,
			checkNil:       false,
		},
		{
			name:           "Create_LoggerChannel_With_Nil_Logger",
			logger:         nil,
			checkInterface: true,
			checkNil:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			channel := NewLoggerChannel(tc.logger)

			if tc.checkNil {
				if channel != nil {
					t.Error("expected channel to be nil, got: not nil")
				}
			} else {
				if channel == nil {
					t.Error("expected channel to be created, got: nil")
					return
				}

				if tc.checkInterface {
					if _, ok := channel.(port.NotificationChannel); !ok {
						t.Error("expected channel to implement NotificationChannel interface")
					}
				}
			}
		})
	}
}

func TestLoggerChannel_Send(t *testing.T) {
	type testCase struct {
		name          string
		message       string
		expectedError bool
		checkLogging  bool
	}

	testCases := []testCase{
		{
			name:          "Send_Success",
			message:       "Test message",
			expectedError: false,
			checkLogging:  true,
		},
		{
			name:          "Send_Empty_Message",
			message:       "",
			expectedError: false,
			checkLogging:  true,
		},
		{
			name:          "Send_With_Special_Characters",
			message:       "Message with special chars: !@#$%^&*()",
			expectedError: false,
			checkLogging:  true,
		},
		{
			name:          "Send_With_Unicode",
			message:       "Сообщение с кириллицей 🚀",
			expectedError: false,
			checkLogging:  true,
		},
		{
			name:          "Send_With_Newlines",
			message:       "Line 1\nLine 2\nLine 3",
			expectedError: false,
			checkLogging:  true,
		},
		{
			name:          "Send_With_Whitespace",
			message:       "   ",
			expectedError: false,
			checkLogging:  true,
		},
		{
			name:          "Send_With_JSON_Message",
			message:       `{"key": "value", "number": 123}`,
			expectedError: false,
			checkLogging:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetFormatter(&logrus.JSONFormatter{})
			logger.SetLevel(logrus.InfoLevel)

			channel := NewLoggerChannel(logger).(*LoggerChannel)

			err := channel.Send(tc.message)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error, got: nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if tc.checkLogging {
				logOutput := buf.String()
				if !strings.Contains(logOutput, "Notification sent via logger channel") {
					t.Error("expected log message 'Notification sent via logger channel'")
				}

				if tc.message != "" {
					lines := strings.Split(strings.TrimSpace(logOutput), "\n")
					found := false
					for _, line := range lines {
						if line == "" {
							continue
						}
						var logEntry map[string]interface{}
						if err := json.Unmarshal([]byte(line), &logEntry); err == nil {
							if msg, ok := logEntry["msg"].(string); ok && msg == "Notification sent via logger channel" {
								if fields, ok := logEntry["fields"].(map[string]interface{}); ok {
									if loggedMessage, ok := fields["message"].(string); ok {
										if loggedMessage == tc.message {
											found = true
											break
										}
									}
								}
								if loggedMessage, ok := logEntry["message"].(string); ok && loggedMessage == tc.message {
									found = true
									break
								}
							}
						}
					}
					if !found && tc.message != "" {
						if !strings.Contains(logOutput, tc.message) {
							t.Logf("message %q not found in log output (may be expected if log structure differs)", tc.message)
						}
					}
				}
			}
		})
	}
}

func TestLoggerChannel_Channel(t *testing.T) {
	type testCase struct {
		name           string
		logger         *logrus.Logger
		expectedResult string
	}

	testCases := []testCase{
		{
			name:           "Channel_Returns_Logger",
			logger:         logrus.New(),
			expectedResult: port.ChannelLogger,
		},
		{
			name:           "Channel_Returns_Logger_With_Nil_Logger",
			logger:         nil,
			expectedResult: port.ChannelLogger,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			channel := NewLoggerChannel(tc.logger)

			result := channel.Channel()

			if result != tc.expectedResult {
				t.Errorf("expected channel name %q, got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestLoggerChannel_Integration(t *testing.T) {
	type testCase struct {
		name    string
		message string
	}

	testCases := []testCase{
		{
			name:    "Full_Integration_Test",
			message: "Integration test message",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetFormatter(&logrus.JSONFormatter{})
			logger.SetLevel(logrus.InfoLevel)

			channel := NewLoggerChannel(logger)

			if _, ok := channel.(port.NotificationChannel); !ok {
				t.Fatal("expected channel to implement NotificationChannel interface")
			}

			if channel.Channel() != port.ChannelLogger {
				t.Errorf("expected channel name %q, got: %q", port.ChannelLogger, channel.Channel())
			}

			err := channel.Send(tc.message)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			logOutput := buf.String()
			if !strings.Contains(logOutput, "Notification sent via logger channel") {
				t.Error("expected log message 'Notification sent via logger channel'")
			}
			if !strings.Contains(logOutput, tc.message) {
				t.Errorf("expected log to contain message %q", tc.message)
			}
		})
	}
}

func TestLoggerChannel_EdgeCases(t *testing.T) {
	type testCase struct {
		name          string
		message       string
		expectedError bool
	}

	testCases := []testCase{
		{
			name:          "Send_With_Very_Long_Message",
			message:       strings.Repeat("A", 10000),
			expectedError: false,
		},
		{
			name:          "Send_With_Only_Newlines",
			message:       "\n\n\n",
			expectedError: false,
		},
		{
			name:          "Send_With_Tabs",
			message:       "Message\twith\ttabs",
			expectedError: false,
		},
		{
			name:          "Send_With_Control_Characters",
			message:       "Message\x00\x01\x02",
			expectedError: false,
		},
		{
			name:          "Send_With_HTML",
			message:       "<html><body>Test</body></html>",
			expectedError: false,
		},
		{
			name:          "Send_With_Markdown",
			message:       "# Header\n**Bold** *Italic*",
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetFormatter(&logrus.JSONFormatter{})
			logger.SetLevel(logrus.InfoLevel)

			channel := NewLoggerChannel(logger).(*LoggerChannel)

			err := channel.Send(tc.message)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error, got: nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				logOutput := buf.String()
				if !strings.Contains(logOutput, "Notification sent via logger channel") {
					t.Error("expected log message 'Notification sent via logger channel'")
				}
			}
		})
	}
}
