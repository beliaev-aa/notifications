package http

import (
	"bytes"
	"errors"
	"github.com/beliaev-aa/notifications/tests/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewHandler(t *testing.T) {
	type testCase struct {
		name     string
		logger   *logrus.Logger
		checkNil bool
	}

	testCases := []testCase{
		{
			name:     "Create_Handler_Success",
			logger:   logrus.New(),
			checkNil: false,
		},
		{
			name:     "Create_Handler_With_Nil_Logger",
			logger:   nil,
			checkNil: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			handler := NewHandler(mockWebhookService, tc.logger)

			if tc.checkNil {
				if handler != nil {
					t.Error("expected handler to be nil, got: not nil")
				}
			} else {
				if handler == nil {
					t.Error("expected handler to be created, got: nil")
					return
				}

				if handler.webhookService == nil {
					t.Error("expected webhookService to be set, got: nil")
				}
				if handler.logger != tc.logger {
					t.Errorf("expected logger to be %v, got: %v", tc.logger, handler.logger)
				}
			}
		})
	}
}

func TestHandler_Health(t *testing.T) {
	type testCase struct {
		name          string
		writeError    error
		expectedError bool
		checkResponse bool
		expectedBody  string
		expectedCode  int
	}

	testCases := []testCase{
		{
			name:          "Health_Success",
			writeError:    nil,
			expectedError: false,
			checkResponse: true,
			expectedBody:  "ok",
			expectedCode:  http.StatusOK,
		},
		{
			name:          "Health_Write_Error",
			writeError:    errors.New("write error"),
			expectedError: true,
			checkResponse: true,
			expectedBody:  "",
			expectedCode:  http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetLevel(logrus.ErrorLevel)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			handler := NewHandler(mockWebhookService, logger)

			req := httptest.NewRequest("GET", "/health", nil)
			recorder := httptest.NewRecorder()

			if tc.writeError != nil {
				errorWriter := &errorResponseWriter{
					ResponseWriter: recorder,
					writeError:     tc.writeError,
				}
				handler.Health(errorWriter, req)
			} else {
				handler.Health(recorder, req)
			}

			if tc.checkResponse {
				if recorder.Code != tc.expectedCode {
					t.Errorf("expected status code %d, got: %d", tc.expectedCode, recorder.Code)
				}

				if tc.expectedBody != "" {
					if recorder.Body.String() != tc.expectedBody {
						t.Errorf("expected body %q, got: %q", tc.expectedBody, recorder.Body.String())
					}
				}
			}

			if tc.expectedError {
				logOutput := buf.String()
				if !strings.Contains(logOutput, "Failed to write health response") {
					t.Error("expected log message about write error")
				}
			}
		})
	}
}

func TestHandler_YoutrackWebhook(t *testing.T) {
	type testCase struct {
		name          string
		requestBody   string
		processError  error
		writeError    error
		expectedError bool
		expectedCode  int
		expectedBody  string
		checkLogging  bool
	}

	testCases := []testCase{
		{
			name:          "YoutrackWebhook_Success",
			requestBody:   `{"test": "data"}`,
			processError:  nil,
			writeError:    nil,
			expectedError: false,
			expectedCode:  http.StatusOK,
			expectedBody:  "ok",
			checkLogging:  true,
		},
		{
			name:          "YoutrackWebhook_ProcessWebhook_Error",
			requestBody:   `{"test": "data"}`,
			processError:  errors.New("process error"),
			writeError:    nil,
			expectedError: true,
			expectedCode:  http.StatusBadRequest,
			expectedBody:  "Failed to process webhook",
			checkLogging:  true,
		},
		{
			name:          "YoutrackWebhook_Write_Error",
			requestBody:   `{"test": "data"}`,
			processError:  nil,
			writeError:    errors.New("write error"),
			expectedError: true,
			expectedCode:  http.StatusInternalServerError,
			expectedBody:  "Failed to write response",
			checkLogging:  true,
		},
		{
			name:          "YoutrackWebhook_With_Empty_Body",
			requestBody:   "",
			processError:  nil,
			writeError:    nil,
			expectedError: false,
			expectedCode:  http.StatusOK,
			expectedBody:  "ok",
			checkLogging:  true,
		},
		{
			name:          "YoutrackWebhook_With_JSON_Body",
			requestBody:   `{"project": {"name": "Test"}, "issue": {"summary": "Test Issue"}}`,
			processError:  nil,
			writeError:    nil,
			expectedError: false,
			expectedCode:  http.StatusOK,
			expectedBody:  "ok",
			checkLogging:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetLevel(logrus.ErrorLevel)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			handler := NewHandler(mockWebhookService, logger)

			req := httptest.NewRequest("POST", "/webhook/youtrack", strings.NewReader(tc.requestBody))
			recorder := httptest.NewRecorder()

			processCall := mockWebhookService.EXPECT().ProcessWebhook(gomock.Any())
			if tc.processError != nil {
				processCall.Return(tc.processError)
			} else {
				processCall.Return(nil)
			}

			if tc.writeError != nil {
				errorWriter := &errorResponseWriter{
					ResponseWriter: recorder,
					writeError:     tc.writeError,
				}
				handler.YoutrackWebhook(errorWriter, req)
			} else {
				handler.YoutrackWebhook(recorder, req)
			}

			if recorder.Code != tc.expectedCode {
				t.Errorf("expected status code %d, got: %d", tc.expectedCode, recorder.Code)
			}

			if tc.expectedBody != "" {
				if !strings.Contains(recorder.Body.String(), tc.expectedBody) {
					t.Errorf("expected body to contain %q, got: %q", tc.expectedBody, recorder.Body.String())
				}
			}

			if tc.checkLogging {
				logOutput := buf.String()
				if tc.processError != nil {
					if !strings.Contains(logOutput, "Failed to process webhook") {
						t.Error("expected log message about process webhook error")
					}
				} else if tc.writeError != nil {
					if !strings.Contains(logOutput, "Failed to write response") {
						t.Error("expected log message about write error")
					}
				}
			}
		})
	}
}

func TestHandler_Integration(t *testing.T) {
	type testCase struct {
		name         string
		method       string
		path         string
		requestBody  string
		expectedCode int
	}

	testCases := []testCase{
		{
			name:         "Health_Endpoint_Integration",
			method:       "GET",
			path:         "/health",
			requestBody:  "",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Webhook_Endpoint_Integration",
			method:       "POST",
			path:         "/webhook/youtrack",
			requestBody:  `{"test": "data"}`,
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			handler := NewHandler(mockWebhookService, logger)

			var req *http.Request
			if tc.requestBody != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.requestBody))
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			recorder := httptest.NewRecorder()

			if tc.method == "GET" {
				handler.Health(recorder, req)
			} else {
				mockWebhookService.EXPECT().ProcessWebhook(gomock.Any()).Return(nil)
				handler.YoutrackWebhook(recorder, req)
			}

			if recorder.Code != tc.expectedCode {
				t.Errorf("expected status code %d, got: %d", tc.expectedCode, recorder.Code)
			}
		})
	}
}

func TestHandler_EdgeCases(t *testing.T) {
	type testCase struct {
		name          string
		requestBody   string
		processError  error
		expectedCode  int
		expectedError bool
	}

	testCases := []testCase{
		{
			name:          "YoutrackWebhook_With_Very_Long_Body",
			requestBody:   strings.Repeat("A", 10000),
			processError:  nil,
			expectedCode:  http.StatusOK,
			expectedError: false,
		},
		{
			name:          "YoutrackWebhook_With_Invalid_JSON",
			requestBody:   `{invalid json}`,
			processError:  errors.New("parse error"),
			expectedCode:  http.StatusBadRequest,
			expectedError: true,
		},
		{
			name:          "YoutrackWebhook_With_Unicode",
			requestBody:   `{"message": "Сообщение с кириллицей 🚀"}`,
			processError:  nil,
			expectedCode:  http.StatusOK,
			expectedError: false,
		},
		{
			name:          "YoutrackWebhook_With_Special_Characters",
			requestBody:   `{"message": "Test & <test> \"quotes\""}`,
			processError:  nil,
			expectedCode:  http.StatusOK,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			handler := NewHandler(mockWebhookService, logger)

			req := httptest.NewRequest("POST", "/webhook/youtrack", strings.NewReader(tc.requestBody))
			recorder := httptest.NewRecorder()

			processCall := mockWebhookService.EXPECT().ProcessWebhook(gomock.Any())
			if tc.processError != nil {
				processCall.Return(tc.processError)
			} else {
				processCall.Return(nil)
			}

			handler.YoutrackWebhook(recorder, req)

			if recorder.Code != tc.expectedCode {
				t.Errorf("expected status code %d, got: %d", tc.expectedCode, recorder.Code)
			}
		})
	}
}

type errorResponseWriter struct {
	http.ResponseWriter
	writeError        error
	writeCalled       bool
	allowHeaderChange bool
}

func (e *errorResponseWriter) Write(b []byte) (int, error) {
	if e.writeError != nil && !e.writeCalled {
		e.writeCalled = true
		e.allowHeaderChange = true
		return 0, e.writeError
	}
	return e.ResponseWriter.Write(b)
}

func (e *errorResponseWriter) WriteHeader(statusCode int) {
	if e.allowHeaderChange {
		e.ResponseWriter.WriteHeader(statusCode)
		return
	}
	if recorder, ok := e.ResponseWriter.(*httptest.ResponseRecorder); ok {
		if recorder.Code == 0 {
			e.ResponseWriter.WriteHeader(statusCode)
		}
	} else {
		e.ResponseWriter.WriteHeader(statusCode)
	}
}
