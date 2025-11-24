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
			service := NewWebhookService(mockSender, tc.logger)

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
		sendError         error
		expectedError     bool
		expectedErrorMsg  string
		checkLogging      bool
		verifySendCalls   bool
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
			name:             "Success_Process_Webhook",
			requestBody:      validJSON,
			parseJSONPayload: validPayload,
			expectedError:    false,
			checkLogging:     true,
			verifySendCalls:  true,
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
			name:             "Error_Sending_Notification_But_Continue",
			requestBody:      validJSON,
			parseJSONPayload: validPayload,
			sendError:        errors.New("send error"),
			expectedError:    false,
			checkLogging:     true,
			verifySendCalls:  true,
		},
		{
			name:              "Error_Closing_Request_Body",
			requestBody:       validJSON,
			requestCloseError: errors.New("close error"),
			parseJSONPayload:  validPayload,
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

					mockParser.EXPECT().NewFormatter().Return(mockFormatter)

					mockFormatter.EXPECT().RegisterChannelFormatter(port.ChannelTelegram, gomock.Any())

					mockFormatter.EXPECT().Format(tc.parseJSONPayload, port.ChannelLogger).Return("formatted for logger")
					mockFormatter.EXPECT().Format(tc.parseJSONPayload, port.ChannelTelegram).Return("formatted for telegram")

					if tc.verifySendCalls {
						sendCall1 := mockSender.EXPECT().Send(port.ChannelLogger, gomock.Any())
						sendCall2 := mockSender.EXPECT().Send(port.ChannelTelegram, gomock.Any())
						if tc.sendError != nil {
							sendCall1.Return(tc.sendError)
							sendCall2.Return(tc.sendError)
						} else {
							sendCall1.Return(nil)
							sendCall2.Return(nil)
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
				}
			}
		})
	}
}

func TestDetermineNotificationChannels(t *testing.T) {
	type testCase struct {
		name             string
		expectedChannels []string
	}

	testCases := []testCase{
		{
			name:             "Returns_Logger_And_Telegram_Channels",
			expectedChannels: []string{port.ChannelLogger, port.ChannelTelegram},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &WebhookService{}

			channels := service.determineNotificationChannels()

			if diff := cmp.Diff(tc.expectedChannels, channels); diff != "" {
				t.Errorf("Unexpected channels (-want +got):\n%s", diff)
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

			mockParser.EXPECT().ParseJSON(gomock.Any()).Return(tc.parseJSONPayload, nil)
			mockParser.EXPECT().NewFormatter().Return(mockFormatter)
			mockFormatter.EXPECT().RegisterChannelFormatter(port.ChannelTelegram, gomock.Any())

			mockFormatter.EXPECT().Format(tc.parseJSONPayload, port.ChannelLogger).Return("formatted for logger")
			mockFormatter.EXPECT().Format(tc.parseJSONPayload, port.ChannelTelegram).Return("formatted for telegram")

			mockSender.EXPECT().Send(port.ChannelLogger, gomock.Any()).
				Do(func(channel string, message string) {
					receivedChannels = append(receivedChannels, channel)
					receivedMessages = append(receivedMessages, message)
				}).
				Return(nil)
			mockSender.EXPECT().Send(port.ChannelTelegram, gomock.Any()).
				Do(func(channel string, message string) {
					receivedChannels = append(receivedChannels, channel)
					receivedMessages = append(receivedMessages, message)
				}).
				Return(nil)

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
				expectedMessages := []string{"formatted for logger", "formatted for telegram"}
				if diff := cmp.Diff(expectedMessages, receivedMessages); diff != "" {
					t.Errorf("Unexpected messages (-want +got):\n%s", diff)
				}
			}
		})
	}
}
