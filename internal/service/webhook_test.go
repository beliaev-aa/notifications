package service

import (
	"bytes"
	"errors"
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"github.com/beliaev-aa/notifications/tests/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockReadCloser struct {
	reader   io.Reader
	closeErr error
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if m.reader == nil {
		return 0, errors.New("read error")
	}
	return m.reader.Read(p)
}

func (m *mockReadCloser) Close() error {
	return m.closeErr
}

func TestNewWebhookService(t *testing.T) {
	type testCase struct {
		name   string
		logger *logrus.Logger
	}

	testCases := []testCase{
		{
			name:   "Create_Service_With_Valid_Parameters",
			logger: logrus.New(),
		},
		{
			name:   "Create_Service_With_Nil_Logger",
			logger: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSender := mocks.NewMockNotificationSender(ctrl)
			mockParser := mocks.NewMockYoutrackParser(ctrl)
			service := NewWebhookService(mockSender, mockParser, tc.logger)

			if service == nil {
				t.Error("expected service to be created, got: nil")
			}

			if _, ok := service.(port.WebhookService); !ok {
				t.Error("expected service to implement WebhookService interface")
			}
		})
	}
}

func TestProcessWebhook(t *testing.T) {
	type testCase struct {
		name              string
		requestBody       string
		requestBodyError  error
		requestCloseError error
		parseJSONError    error
		parseJSONPayload  *parser.YoutrackWebhookPayload
		projectName       string
		projectAllowed    bool
		allowedChannels   []string
		telegramChatID    string
		hasTelegramChatID bool
		vkteamsChatID     string
		hasVKTeamsChatID  bool
		sendError         error
		expectedError     bool
		expectedErrorMsg  string
		checkLogging      bool
		verifySendCalls   bool
	}

	validJSON := `{"project":{"name":"TestProject"},"issue":{"summary":"Test Issue","url":"https://test.com"}}`
	fieldName := "TestProject"
	validPayload := &parser.YoutrackWebhookPayload{
		Project: &parser.YoutrackFieldValue{Name: &fieldName},
		Issue: parser.YoutrackIssue{
			Summary: "Test Issue",
			URL:     "https://test.com",
		},
	}

	testCases := []testCase{
		{
			name:              "Success_Process_Webhook",
			requestBody:       validJSON,
			parseJSONPayload:  validPayload,
			projectName:       "TestProject",
			projectAllowed:    true,
			allowedChannels:   []string{port.ChannelLogger, port.ChannelTelegram},
			telegramChatID:    "test_chat_id",
			hasTelegramChatID: true,
			expectedError:     false,
			checkLogging:      true,
			verifySendCalls:   true,
		},
		{
			name:             "Error_Reading_Request_Body",
			requestBodyError: errors.New("read error"),
			expectedError:    true,
			expectedErrorMsg: "read error",
		},
		{
			name:             "Error_Parsing_JSON",
			requestBody:      "invalid json",
			parseJSONError:   errors.New("failed to unmarshal webhook payload"),
			expectedError:    true,
			expectedErrorMsg: "failed to unmarshal webhook payload",
			checkLogging:     true,
		},
		{
			name:             "Project_Not_Allowed_Ignores_Notification",
			requestBody:      validJSON,
			parseJSONPayload: validPayload,
			projectName:      "TestProject",
			projectAllowed:   false,
			expectedError:    false,
			checkLogging:     true,
			verifySendCalls:  false,
		},
		{
			name:        "Project_Name_Empty_Ignores_Notification",
			requestBody: `{"project":null,"issue":{"summary":"Test Issue","url":"https://test.com"}}`,
			parseJSONPayload: &parser.YoutrackWebhookPayload{
				Project: nil,
				Issue: parser.YoutrackIssue{
					Summary: "Test Issue",
					URL:     "https://test.com",
				},
			},
			expectedError:   false,
			checkLogging:    true,
			verifySendCalls: false,
		},
		{
			name:              "Error_Sending_Notification_But_Continue",
			requestBody:       validJSON,
			parseJSONPayload:  validPayload,
			projectName:       "TestProject",
			projectAllowed:    true,
			allowedChannels:   []string{port.ChannelLogger, port.ChannelTelegram},
			telegramChatID:    "test_chat_id",
			hasTelegramChatID: true,
			sendError:         errors.New("send error"),
			expectedError:     false,
			checkLogging:      true,
			verifySendCalls:   true,
		},
		{
			name:              "Error_Closing_Request_Body",
			requestBody:       validJSON,
			requestCloseError: errors.New("close error"),
			parseJSONPayload:  validPayload,
			projectName:       "TestProject",
			projectAllowed:    true,
			allowedChannels:   []string{port.ChannelLogger, port.ChannelTelegram},
			telegramChatID:    "test_chat_id",
			hasTelegramChatID: true,
			expectedError:     false,
			checkLogging:      true,
			verifySendCalls:   true,
		},
		{
			name:              "Project_With_Only_Logger_Channel",
			requestBody:       validJSON,
			parseJSONPayload:  validPayload,
			projectName:       "TestProject",
			projectAllowed:    true,
			allowedChannels:   []string{port.ChannelLogger},
			telegramChatID:    "",
			hasTelegramChatID: false,
			expectedError:     false,
			checkLogging:      true,
			verifySendCalls:   true,
		},
		{
			name:             "Project_With_Empty_Channels_Ignores",
			requestBody:      validJSON,
			parseJSONPayload: validPayload,
			projectName:      "TestProject",
			projectAllowed:   true,
			allowedChannels:  []string{},
			expectedError:    false,
			checkLogging:     true,
			verifySendCalls:  false,
		},
		{
			name:              "Telegram_Chat_ID_Not_Found_Skips_Telegram_But_Sends_To_Other_Channels",
			requestBody:       validJSON,
			parseJSONPayload:  validPayload,
			projectName:       "TestProject",
			projectAllowed:    true,
			allowedChannels:   []string{port.ChannelTelegram, port.ChannelLogger},
			telegramChatID:    "",
			hasTelegramChatID: false,
			expectedError:     false,
			checkLogging:      true,
			verifySendCalls:   true,
		},
		{
			name:              "Telegram_Chat_ID_Empty_String_Skips_Telegram",
			requestBody:       validJSON,
			parseJSONPayload:  validPayload,
			projectName:       "TestProject",
			projectAllowed:    true,
			allowedChannels:   []string{port.ChannelTelegram, port.ChannelLogger},
			telegramChatID:    "",
			hasTelegramChatID: true,
			expectedError:     false,
			checkLogging:      true,
			verifySendCalls:   true,
		},
		{
			name:              "Telegram_Chat_ID_Not_Found_Only_Telegram_Channel",
			requestBody:       validJSON,
			parseJSONPayload:  validPayload,
			projectName:       "TestProject",
			projectAllowed:    true,
			allowedChannels:   []string{port.ChannelTelegram},
			telegramChatID:    "",
			hasTelegramChatID: false,
			expectedError:     false,
			checkLogging:      true,
			verifySendCalls:   true,
		},
		{
			name:             "VKTeams_Chat_ID_Not_Found_Skips_VKTeams_But_Sends_To_Other_Channels",
			requestBody:      validJSON,
			parseJSONPayload: validPayload,
			projectName:      "TestProject",
			projectAllowed:   true,
			allowedChannels:  []string{port.ChannelVKTeams, port.ChannelLogger},
			vkteamsChatID:    "",
			hasVKTeamsChatID: false,
			expectedError:    false,
			checkLogging:     true,
			verifySendCalls:  true,
		},
		{
			name:             "VKTeams_Chat_ID_Empty_String_Skips_VKTeams",
			requestBody:      validJSON,
			parseJSONPayload: validPayload,
			projectName:      "TestProject",
			projectAllowed:   true,
			allowedChannels:  []string{port.ChannelVKTeams, port.ChannelLogger},
			vkteamsChatID:    "",
			hasVKTeamsChatID: true,
			expectedError:    false,
			checkLogging:     true,
			verifySendCalls:  true,
		},
		{
			name:             "VKTeams_Chat_ID_Not_Found_Only_VKTeams_Channel",
			requestBody:      validJSON,
			parseJSONPayload: validPayload,
			projectName:      "TestProject",
			projectAllowed:   true,
			allowedChannels:  []string{port.ChannelVKTeams},
			vkteamsChatID:    "",
			hasVKTeamsChatID: false,
			expectedError:    false,
			checkLogging:     true,
			verifySendCalls:  true,
		},
		{
			name:             "Project_With_VKTeams_Channel",
			requestBody:      validJSON,
			parseJSONPayload: validPayload,
			projectName:      "TestProject",
			projectAllowed:   true,
			allowedChannels:  []string{port.ChannelVKTeams, port.ChannelLogger},
			vkteamsChatID:    "test_vkteams_chat_id",
			hasVKTeamsChatID: true,
			expectedError:    false,
			checkLogging:     true,
			verifySendCalls:  true,
		},
		{
			name:             "Project_With_Only_VKTeams_Channel",
			requestBody:      validJSON,
			parseJSONPayload: validPayload,
			projectName:      "TestProject",
			projectAllowed:   true,
			allowedChannels:  []string{port.ChannelVKTeams},
			vkteamsChatID:    "test_vkteams_chat_id",
			hasVKTeamsChatID: true,
			expectedError:    false,
			checkLogging:     true,
			verifySendCalls:  true,
		},
		{
			name:              "Project_With_Both_Telegram_And_VKTeams",
			requestBody:       validJSON,
			parseJSONPayload:  validPayload,
			projectName:       "TestProject",
			projectAllowed:    true,
			allowedChannels:   []string{port.ChannelTelegram, port.ChannelVKTeams, port.ChannelLogger},
			telegramChatID:    "test_telegram_chat_id",
			hasTelegramChatID: true,
			vkteamsChatID:     "test_vkteams_chat_id",
			hasVKTeamsChatID:  true,
			expectedError:     false,
			checkLogging:      true,
			verifySendCalls:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetFormatter(&logrus.JSONFormatter{})
			logger.SetLevel(logrus.DebugLevel)

			mockSender := mocks.NewMockNotificationSender(ctrl)
			mockParser := mocks.NewMockYoutrackParser(ctrl)
			mockFormatter := mocks.NewMockYoutrackFormatter(ctrl)

			var requestBody io.ReadCloser
			if tc.requestBodyError != nil {
				requestBody = &mockReadCloser{
					reader: nil,
				}
			} else {
				requestBody = &mockReadCloser{
					reader:   strings.NewReader(tc.requestBody),
					closeErr: tc.requestCloseError,
				}
			}

			req, err := http.NewRequest("POST", "/webhook", requestBody)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			if tc.requestBodyError == nil {
				parseCall := mockParser.EXPECT().ParseJSON(gomock.Any())
				if tc.parseJSONError != nil {
					parseCall.Return(nil, tc.parseJSONError)
				} else {
					parseCall.Return(tc.parseJSONPayload, nil)

					if tc.parseJSONPayload != nil {
						mockParser.EXPECT().GetAllowedChannels(tc.parseJSONPayload).Return(tc.allowedChannels)
					}

					if tc.verifySendCalls && len(tc.allowedChannels) > 0 {
						mockParser.EXPECT().NewFormatter().Return(mockFormatter)
						mockFormatter.EXPECT().RegisterChannelFormatter(port.ChannelTelegram, gomock.Any())
						mockFormatter.EXPECT().RegisterChannelFormatter(port.ChannelVKTeams, gomock.Any())

						projectName := ""
						if tc.parseJSONPayload != nil && tc.parseJSONPayload.Project != nil && tc.parseJSONPayload.Project.Name != nil {
							projectName = *tc.parseJSONPayload.Project.Name
						}
						if projectName == "" {
							projectName = tc.projectName
						}

						for _, channel := range tc.allowedChannels {
							mockFormatter.EXPECT().Format(tc.parseJSONPayload, channel).Return("formatted for " + channel)

							chatID := ""
							shouldSkip := false
							if channel == port.ChannelTelegram {
								if projectName != "" {
									mockParser.EXPECT().GetTelegramChatID(projectName).Return(tc.telegramChatID, tc.hasTelegramChatID)
								}
								if !tc.hasTelegramChatID || tc.telegramChatID == "" {
									shouldSkip = true
								} else {
									chatID = tc.telegramChatID
								}
							} else if channel == port.ChannelVKTeams {
								if projectName != "" {
									mockParser.EXPECT().GetVKTeamsChatID(projectName).Return(tc.vkteamsChatID, tc.hasVKTeamsChatID)
								}
								if !tc.hasVKTeamsChatID || tc.vkteamsChatID == "" {
									shouldSkip = true
								} else {
									chatID = tc.vkteamsChatID
								}
							}

							if shouldSkip {
								continue
							}

							sendCall := mockSender.EXPECT().Send(channel, chatID, gomock.Any())
							if tc.sendError != nil {
								sendCall.Return(tc.sendError)
							} else {
								sendCall.Return(nil)
							}
						}
					}
				}
			}

			service := &WebhookService{
				notificationSender: mockSender,
				youtrackParser:     mockParser,
				logger:             logger,
			}

			err = service.ProcessWebhook(req)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error, got: nil")
				} else if tc.expectedErrorMsg != "" && !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("expected error message to contain %q, got: %v", tc.expectedErrorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if tc.checkLogging {
				logOutput := buf.String()
				if tc.requestBodyError != nil {
					if !strings.Contains(logOutput, "Failed to read request body") {
						t.Error("expected log message about reading request body error")
					}
				} else if tc.parseJSONError != nil {
					if !strings.Contains(logOutput, "Failed to parse YouTrack data") {
						t.Error("expected log message about parsing error")
					}
				} else {
					if !strings.Contains(logOutput, "Webhook received") && !strings.Contains(logOutput, "webhook") {
					}
					if strings.Contains(tc.name, "Telegram_Chat_ID_Not_Found") || strings.Contains(tc.name, "Telegram_Chat_ID_Empty") {
						if !strings.Contains(logOutput, "Telegram chat ID not found for project") && !strings.Contains(logOutput, "skipping notification") {
							t.Error("expected log message about Telegram chat ID not found")
						}
					}
					if strings.Contains(tc.name, "VKTeams_Chat_ID_Not_Found") || strings.Contains(tc.name, "VKTeams_Chat_ID_Empty") {
						if !strings.Contains(logOutput, "VK Teams chat ID not found for project") && !strings.Contains(logOutput, "skipping notification") {
							t.Error("expected log message about VK Teams chat ID not found")
						}
					}
				}
			}
		})
	}
}

func TestProcessWebhook_Integration(t *testing.T) {
	type testCase struct {
		name             string
		requestBody      string
		parseJSONPayload *parser.YoutrackWebhookPayload
		expectedChannels []string
		checkFormatter   bool
	}

	validJSON := `{"project":{"name":"Test"},"issue":{"summary":"Test Issue","url":"https://test.com"}}`
	fieldName := "Test"
	validPayload := &parser.YoutrackWebhookPayload{
		Project: &parser.YoutrackFieldValue{Name: &fieldName},
		Issue: parser.YoutrackIssue{
			Summary: "Test Issue",
			URL:     "https://test.com",
		},
	}

	testCases := []testCase{
		{
			name:             "Full_Integration_Test",
			requestBody:      validJSON,
			parseJSONPayload: validPayload,
			expectedChannels: []string{port.ChannelLogger, port.ChannelTelegram},
			checkFormatter:   true,
		},
		{
			name:             "Full_Integration_Test_With_VKTeams",
			requestBody:      validJSON,
			parseJSONPayload: validPayload,
			expectedChannels: []string{port.ChannelLogger, port.ChannelVKTeams},
			checkFormatter:   true,
		},
		{
			name:             "Full_Integration_Test_With_Both_Telegram_And_VKTeams",
			requestBody:      validJSON,
			parseJSONPayload: validPayload,
			expectedChannels: []string{port.ChannelLogger, port.ChannelTelegram, port.ChannelVKTeams},
			checkFormatter:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetFormatter(&logrus.JSONFormatter{})
			logger.SetLevel(logrus.DebugLevel)

			var receivedChannels []string
			var receivedMessages []string

			mockSender := mocks.NewMockNotificationSender(ctrl)
			mockParser := mocks.NewMockYoutrackParser(ctrl)
			mockFormatter := mocks.NewMockYoutrackFormatter(ctrl)

			requestBody := io.NopCloser(strings.NewReader(tc.requestBody))
			req, err := http.NewRequest("POST", "/webhook", requestBody)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			projectName := "Test"
			if tc.parseJSONPayload != nil && tc.parseJSONPayload.Project != nil {
				if tc.parseJSONPayload.Project.Name != nil {
					projectName = *tc.parseJSONPayload.Project.Name
				}
			}

			mockParser.EXPECT().ParseJSON(gomock.Any()).Return(tc.parseJSONPayload, nil)
			mockParser.EXPECT().GetAllowedChannels(tc.parseJSONPayload).Return(tc.expectedChannels)

			mockParser.EXPECT().NewFormatter().Return(mockFormatter)
			mockFormatter.EXPECT().RegisterChannelFormatter(port.ChannelTelegram, gomock.Any())
			mockFormatter.EXPECT().RegisterChannelFormatter(port.ChannelVKTeams, gomock.Any())

			hasTelegram := false
			hasVKTeams := false
			for _, ch := range tc.expectedChannels {
				if ch == port.ChannelTelegram {
					hasTelegram = true
				}
				if ch == port.ChannelVKTeams {
					hasVKTeams = true
				}
			}

			for _, channel := range tc.expectedChannels {
				mockFormatter.EXPECT().Format(tc.parseJSONPayload, channel).Return("formatted for " + channel)
			}

			if hasTelegram {
				mockParser.EXPECT().GetTelegramChatID(projectName).Return("test_chat_id", true)
			}
			if hasVKTeams {
				mockParser.EXPECT().GetVKTeamsChatID(projectName).Return("test_vkteams_chat_id", true)
			}

			for _, channel := range tc.expectedChannels {
				chatID := ""
				if channel == port.ChannelTelegram {
					chatID = "test_chat_id"
				} else if channel == port.ChannelVKTeams {
					chatID = "test_vkteams_chat_id"
				}

				mockSender.EXPECT().Send(channel, chatID, gomock.Any()).
					Do(func(ch string, cID string, msg string) {
						receivedChannels = append(receivedChannels, ch)
						receivedMessages = append(receivedMessages, msg)
					}).
					Return(nil)
			}

			service := &WebhookService{
				notificationSender: mockSender,
				youtrackParser:     mockParser,
				logger:             logger,
			}

			err = service.ProcessWebhook(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.expectedChannels, receivedChannels); diff != "" {
				t.Errorf("Unexpected channels (-want +got):\n%s", diff)
			}

			if tc.checkFormatter {
				if len(receivedMessages) != len(tc.expectedChannels) {
					t.Errorf("expected %d formatted messages, got: %d", len(tc.expectedChannels), len(receivedMessages))
				}
				expectedMessages := make([]string, 0, len(tc.expectedChannels))
				for _, ch := range tc.expectedChannels {
					expectedMessages = append(expectedMessages, "formatted for "+ch)
				}
				if diff := cmp.Diff(expectedMessages, receivedMessages); diff != "" {
					t.Errorf("Unexpected messages (-want +got):\n%s", diff)
				}
			}
		})
	}
}
