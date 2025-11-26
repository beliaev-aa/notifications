package http

import (
	"github.com/beliaev-aa/notifications/tests/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRouter(t *testing.T) {
	type testCase struct {
		name           string
		logger         *logrus.Logger
		checkNil       bool
		checkRoutes    bool
		expectedRoutes []struct {
			method string
			path   string
		}
	}

	testCases := []testCase{
		{
			name:        "Create_Router_Success",
			logger:      logrus.New(),
			checkNil:    false,
			checkRoutes: true,
			expectedRoutes: []struct {
				method string
				path   string
			}{
				{method: "GET", path: "/health"},
				{method: "POST", path: "/webhook/youtrack"},
			},
		},
		{
			name:        "Create_Router_With_Nil_Logger",
			logger:      nil,
			checkNil:    false,
			checkRoutes: true,
			expectedRoutes: []struct {
				method string
				path   string
			}{
				{method: "GET", path: "/health"},
				{method: "POST", path: "/webhook/youtrack"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			router := NewRouter(mockWebhookService, tc.logger)

			if tc.checkNil {
				if router != nil {
					t.Error("expected router to be nil, got: not nil")
				}
			} else {
				if router == nil {
					t.Error("expected router to be created, got: nil")
					return
				}

				if tc.checkRoutes {
					for _, route := range tc.expectedRoutes {
						req := httptest.NewRequest(route.method, route.path, nil)
						recorder := httptest.NewRecorder()

						if route.method == "POST" && route.path == "/webhook/youtrack" {
							mockWebhookService.EXPECT().ProcessWebhook(gomock.Any()).Return(nil)
						}

						router.ServeHTTP(recorder, req)

						if recorder.Code == http.StatusNotFound {
							t.Errorf("expected route %s %s to be registered, got: 404", route.method, route.path)
						}
					}
				}
			}
		})
	}
}

func TestNewRouter_Routes(t *testing.T) {
	type testCase struct {
		name         string
		method       string
		path         string
		requestBody  string
		expectedCode int
		checkBody    bool
		expectedBody string
	}

	testCases := []testCase{
		{
			name:         "Route_Health_GET",
			method:       "GET",
			path:         "/health",
			requestBody:  "",
			expectedCode: http.StatusOK,
			checkBody:    true,
			expectedBody: "ok",
		},
		{
			name:         "Route_Webhook_POST_Success",
			method:       "POST",
			path:         "/webhook/youtrack",
			requestBody:  `{"test": "data"}`,
			expectedCode: http.StatusOK,
			checkBody:    true,
			expectedBody: "ok",
		},
		{
			name:         "Route_Webhook_POST_With_Empty_Body",
			method:       "POST",
			path:         "/webhook/youtrack",
			requestBody:  "",
			expectedCode: http.StatusOK,
			checkBody:    true,
			expectedBody: "ok",
		},
		{
			name:         "Route_Webhook_POST_With_JSON",
			method:       "POST",
			path:         "/webhook/youtrack",
			requestBody:  `{"project": {"name": "Test"}, "issue": {"summary": "Test Issue"}}`,
			expectedCode: http.StatusOK,
			checkBody:    true,
			expectedBody: "ok",
		},
		{
			name:         "Route_Health_Wrong_Method",
			method:       "POST",
			path:         "/health",
			requestBody:  "",
			expectedCode: http.StatusMethodNotAllowed,
			checkBody:    false,
		},
		{
			name:         "Route_Webhook_Wrong_Method",
			method:       "GET",
			path:         "/webhook/youtrack",
			requestBody:  "",
			expectedCode: http.StatusMethodNotAllowed,
			checkBody:    false,
		},
		{
			name:         "Route_NonExistent_Path",
			method:       "GET",
			path:         "/nonexistent",
			requestBody:  "",
			expectedCode: http.StatusNotFound,
			checkBody:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			router := NewRouter(mockWebhookService, logger)

			var req *http.Request
			if tc.requestBody != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.requestBody))
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			recorder := httptest.NewRecorder()

			if tc.method == "POST" && tc.path == "/webhook/youtrack" {
				mockWebhookService.EXPECT().ProcessWebhook(gomock.Any()).Return(nil)
			}

			router.ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedCode {
				t.Errorf("expected status code %d, got: %d", tc.expectedCode, recorder.Code)
			}

			if tc.checkBody {
				if recorder.Body.String() != tc.expectedBody {
					t.Errorf("expected body %q, got: %q", tc.expectedBody, recorder.Body.String())
				}
			}
		})
	}
}

func TestNewRouter_Integration(t *testing.T) {
	type testCase struct {
		name         string
		method       string
		path         string
		requestBody  string
		processError error
		expectedCode int
		checkBody    bool
		expectedBody string
	}

	testCases := []testCase{
		{
			name:         "Integration_Health_Check",
			method:       "GET",
			path:         "/health",
			requestBody:  "",
			processError: nil,
			expectedCode: http.StatusOK,
			checkBody:    true,
			expectedBody: "ok",
		},
		{
			name:         "Integration_Webhook_Success",
			method:       "POST",
			path:         "/webhook/youtrack",
			requestBody:  `{"project": {"name": "TestProject"}}`,
			processError: nil,
			expectedCode: http.StatusOK,
			checkBody:    true,
			expectedBody: "ok",
		},
		{
			name:         "Integration_Webhook_Process_Error",
			method:       "POST",
			path:         "/webhook/youtrack",
			requestBody:  `{"project": {"name": "TestProject"}}`,
			processError: http.ErrBodyReadAfterClose,
			expectedCode: http.StatusBadRequest,
			checkBody:    true,
			expectedBody: "Failed to process webhook",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var buf strings.Builder
			logger := logrus.New()
			logger.SetOutput(&buf)
			logger.SetLevel(logrus.ErrorLevel)

			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			router := NewRouter(mockWebhookService, logger)

			var req *http.Request
			if tc.requestBody != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.requestBody))
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			recorder := httptest.NewRecorder()

			if tc.method == "POST" {
				processCall := mockWebhookService.EXPECT().ProcessWebhook(gomock.Any())
				if tc.processError != nil {
					processCall.Return(tc.processError)
				} else {
					processCall.Return(nil)
				}
			}

			router.ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedCode {
				t.Errorf("expected status code %d, got: %d", tc.expectedCode, recorder.Code)
			}

			if tc.checkBody {
				if !strings.Contains(recorder.Body.String(), tc.expectedBody) {
					t.Errorf("expected body to contain %q, got: %q", tc.expectedBody, recorder.Body.String())
				}
			}
		})
	}
}

func TestNewRouter_EdgeCases(t *testing.T) {
	type testCase struct {
		name         string
		method       string
		path         string
		requestBody  string
		expectedCode int
	}

	testCases := []testCase{
		{
			name:         "EdgeCase_Health_With_Trailing_Slash",
			method:       "GET",
			path:         "/health/",
			requestBody:  "",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "EdgeCase_Webhook_With_Trailing_Slash",
			method:       "POST",
			path:         "/webhook/youtrack/",
			requestBody:  `{"test": "data"}`,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "EdgeCase_Webhook_With_Query_Params",
			method:       "POST",
			path:         "/webhook/youtrack?param=value",
			requestBody:  `{"test": "data"}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "EdgeCase_Health_With_Query_Params",
			method:       "GET",
			path:         "/health?check=true",
			requestBody:  "",
			expectedCode: http.StatusOK,
		},
		{
			name:         "EdgeCase_Webhook_With_Very_Long_Body",
			method:       "POST",
			path:         "/webhook/youtrack",
			requestBody:  strings.Repeat("A", 10000),
			expectedCode: http.StatusOK,
		},
		{
			name:         "EdgeCase_Webhook_With_Unicode",
			method:       "POST",
			path:         "/webhook/youtrack",
			requestBody:  `{"message": "Сообщение с кириллицей 🚀"}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "EdgeCase_Webhook_With_Special_Characters",
			method:       "POST",
			path:         "/webhook/youtrack",
			requestBody:  `{"message": "Test & <test> \"quotes\""}`,
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
			router := NewRouter(mockWebhookService, logger)

			var req *http.Request
			if tc.requestBody != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.requestBody))
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			recorder := httptest.NewRecorder()

			if tc.method == "POST" && tc.expectedCode == http.StatusOK {
				mockWebhookService.EXPECT().ProcessWebhook(gomock.Any()).Return(nil)
			}

			router.ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedCode {
				t.Errorf("expected status code %d, got: %d", tc.expectedCode, recorder.Code)
			}
		})
	}
}
