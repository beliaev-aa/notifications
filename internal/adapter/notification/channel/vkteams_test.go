package channel

import (
	"errors"
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/beliaev-aa/notifications/tests/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewVKTeamsChannel(t *testing.T) {
	type testCase struct {
		name           string
		cfg            config.VKTeamsConfig
		logger         *logrus.Logger
		checkInterface bool
		checkNil       bool
	}

	testCases := []testCase{
		{
			name: "Create_VKTeamsChannel_Success",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			logger:         logrus.New(),
			checkInterface: true,
			checkNil:       false,
		},
		{
			name: "Create_VKTeamsChannel_With_Empty_Token",
			cfg: config.VKTeamsConfig{
				BotToken: "",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			logger:         logrus.New(),
			checkInterface: true,
			checkNil:       false,
		},
		{
			name: "Create_VKTeamsChannel_With_Valid_Config",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			logger:         logrus.New(),
			checkInterface: true,
			checkNil:       false,
		},
		{
			name: "Create_VKTeamsChannel_With_Nil_Logger",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			logger:         nil,
			checkInterface: true,
			checkNil:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			channel := NewVKTeamsChannel(tc.cfg, tc.logger, mockHTTPClient)

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

func TestVKTeamsChannel_Send(t *testing.T) {
	type testCase struct {
		name             string
		cfg              config.VKTeamsConfig
		chatID           string
		message          string
		httpResponse     *http.Response
		httpError        error
		expectedError    bool
		expectedErrorMsg string
		checkLogging     bool
	}

	successResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"ok": true}`)),
	}

	errorResponse := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader(`{"ok": false, "description": "Bad Request"}`)),
	}

	testCases := []testCase{
		{
			name: "Send_Success",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:        "test_chat_id",
			message:       "Test message",
			httpResponse:  successResponse,
			httpError:     nil,
			expectedError: false,
			checkLogging:  true,
		},
		{
			name: "Send_With_Empty_ChatID",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:           "",
			message:          "Test message",
			httpResponse:     nil,
			httpError:        nil,
			expectedError:    true,
			expectedErrorMsg: "vkteams chat ID is not configured",
		},
		{
			name: "Send_HTTP_Client_Error",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:           "test_chat_id",
			message:          "Test message",
			httpResponse:     nil,
			httpError:        errors.New("network error"),
			expectedError:    true,
			expectedErrorMsg: "failed to send message",
		},
		{
			name: "Send_HTTP_Error_Response",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:           "test_chat_id",
			message:          "Test message",
			httpResponse:     errorResponse,
			httpError:        nil,
			expectedError:    true,
			expectedErrorMsg: "vkteams API error: status 400",
		},
		{
			name: "Send_With_Special_Characters",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:        "test_chat_id",
			message:       "Message with special chars: !@#$%^&*()",
			httpResponse:  successResponse,
			httpError:     nil,
			expectedError: false,
		},
		{
			name: "Send_With_Unicode",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:        "test_chat_id",
			message:       "Сообщение с кириллицей 🚀",
			httpResponse:  successResponse,
			httpError:     nil,
			expectedError: false,
		},
		{
			name: "Send_With_Newlines",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:        "test_chat_id",
			message:       "Line 1\nLine 2\nLine 3",
			httpResponse:  successResponse,
			httpError:     nil,
			expectedError: false,
		},
		{
			name: "Send_With_Empty_Response_Body",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:  "test_chat_id",
			message: "Test message",
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
			},
			httpError:     nil,
			expectedError: false,
		},
		{
			name: "Send_With_500_Error",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:  "test_chat_id",
			message: "Test message",
			httpResponse: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`{"ok": false}`)),
			},
			httpError:        nil,
			expectedError:    true,
			expectedErrorMsg: "vkteams API error: status 500",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

			channel := NewVKTeamsChannel(tc.cfg, logger, mockHTTPClient).(*VKTeamsChannel)

			if tc.cfg.BotToken != "" {
				if tc.httpError != nil {
					mockHTTPClient.EXPECT().Do(gomock.Any()).Return(nil, tc.httpError)
				} else if tc.httpResponse != nil {
					bodyContent := ""
					if tc.httpResponse.Body != nil {
						bodyBytes, _ := io.ReadAll(tc.httpResponse.Body)
						bodyContent = string(bodyBytes)
						err := tc.httpResponse.Body.Close()
						if err != nil {
							return
						}
					}

					mockHTTPClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
						expectedPath := "/messages/sendText"
						if !strings.HasSuffix(req.URL.Path, expectedPath) {
							t.Errorf("expected path to end with %q, got: %q", expectedPath, req.URL.Path)
						}

						if req.Method != "GET" {
							t.Errorf("expected GET method, got: %s", req.Method)
						}

						query := req.URL.Query()
						if query.Get("token") != tc.cfg.BotToken {
							t.Errorf("expected token %q, got: %v", tc.cfg.BotToken, query.Get("token"))
						}
						if query.Get("chatId") != tc.chatID {
							t.Errorf("expected chatId %q, got: %v", tc.chatID, query.Get("chatId"))
						}
						if query.Get("text") != tc.message {
							t.Errorf("expected text %q, got: %v", tc.message, query.Get("text"))
						}
						if query.Get("parseMode") != "MarkdownV2" {
							t.Errorf("expected parseMode %q, got: %v", "MarkdownV2", query.Get("parseMode"))
						}

						response := &http.Response{
							StatusCode: tc.httpResponse.StatusCode,
							Body:       io.NopCloser(strings.NewReader(bodyContent)),
						}
						return response, nil
					})
				}
			}

			err := channel.Send(tc.chatID, tc.message)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error, got: nil")
				} else if tc.expectedErrorMsg != "" && !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("expected error message to contain %q, got: %q", tc.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestVKTeamsChannel_Channel(t *testing.T) {
	type testCase struct {
		name           string
		cfg            config.VKTeamsConfig
		expectedResult string
	}

	testCases := []testCase{
		{
			name: "Channel_Returns_VKTeams",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			expectedResult: port.ChannelVKTeams,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)
			channel := NewVKTeamsChannel(tc.cfg, logrus.New(), mockHTTPClient)

			result := channel.Channel()

			if result != tc.expectedResult {
				t.Errorf("expected channel name %q, got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestVKTeamsChannel_Integration(t *testing.T) {
	type testCase struct {
		name    string
		cfg     config.VKTeamsConfig
		chatID  string
		message string
	}

	testCases := []testCase{
		{
			name: "Full_Integration_Test",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:  "test_chat_id",
			message: "Integration test message",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

			channel := NewVKTeamsChannel(tc.cfg, logger, mockHTTPClient)

			if _, ok := channel.(port.NotificationChannel); !ok {
				t.Fatal("expected channel to implement NotificationChannel interface")
			}

			if channel.Channel() != port.ChannelVKTeams {
				t.Errorf("expected channel name %q, got: %q", port.ChannelVKTeams, channel.Channel())
			}

			mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok": true}`)),
			}, nil)

			err := channel.Send(tc.chatID, tc.message)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestVKTeamsChannel_EdgeCases(t *testing.T) {
	type testCase struct {
		name          string
		cfg           config.VKTeamsConfig
		chatID        string
		message       string
		httpResponse  *http.Response
		httpError     error
		expectedError bool
	}

	testCases := []testCase{
		{
			name: "Send_With_Very_Long_Message",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:  "test_chat_id",
			message: strings.Repeat("A", 10000),
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok": true}`)),
			},
			httpError:     nil,
			expectedError: false,
		},
		{
			name: "Send_With_JSON_In_Message",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:  "test_chat_id",
			message: `{"key": "value", "number": 123}`,
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok": true}`)),
			},
			httpError:     nil,
			expectedError: false,
		},
		{
			name: "Send_With_Timeout_Error",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:        "test_chat_id",
			message:       "Test message",
			httpResponse:  nil,
			httpError:     errors.New("context deadline exceeded"),
			expectedError: true,
		},
		{
			name: "Send_With_Connection_Error",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:        "test_chat_id",
			message:       "Test message",
			httpResponse:  nil,
			httpError:     errors.New("connection refused"),
			expectedError: true,
		},
		{
			name: "Send_With_403_Forbidden",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:  "test_chat_id",
			message: "Test message",
			httpResponse: &http.Response{
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(strings.NewReader(`{"ok": false, "description": "Forbidden"}`)),
			},
			httpError:     nil,
			expectedError: true,
		},
		{
			name: "Send_With_404_Not_Found",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:  "test_chat_id",
			message: "Test message",
			httpResponse: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(`{"ok": false, "description": "Not Found"}`)),
			},
			httpError:     nil,
			expectedError: true,
		},
		{
			name: "Send_With_Read_Body_Error",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:  "test_chat_id",
			message: "Test message",
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &errorReadCloser{readError: errors.New("read error"), reader: strings.NewReader("test")},
			},
			httpError:     nil,
			expectedError: false,
		},
		{
			name: "Send_With_Close_Body_Error",
			cfg: config.VKTeamsConfig{
				BotToken: "test_token",
				Timeout:  10,
				ApiUrl:   "https://api.example.com/bot/v1",
			},
			chatID:  "test_chat_id",
			message: "Test message",
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &errorReadCloser{closeError: errors.New("close error"), reader: strings.NewReader(`{"ok": true}`)},
			},
			httpError:     nil,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

			channel := NewVKTeamsChannel(tc.cfg, logger, mockHTTPClient).(*VKTeamsChannel)

			if tc.httpError != nil {
				mockHTTPClient.EXPECT().Do(gomock.Any()).Return(nil, tc.httpError)
			} else if tc.httpResponse != nil {
				bodyContent := ""
				if tc.httpResponse.Body != nil {
					bodyBytes, _ := io.ReadAll(tc.httpResponse.Body)
					bodyContent = string(bodyBytes)
					err := tc.httpResponse.Body.Close()
					if err != nil {
						return
					}
				}

				mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
					StatusCode: tc.httpResponse.StatusCode,
					Body:       io.NopCloser(strings.NewReader(bodyContent)),
				}, nil)
			}

			err := channel.Send(tc.chatID, tc.message)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error, got: nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
