package http

import (
	"context"
	"github.com/beliaev-aa/notifications/internal/config"
	"github.com/beliaev-aa/notifications/tests/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	type testCase struct {
		name     string
		cfg      *config.HTTPConfig
		logger   *logrus.Logger
		checkNil bool
	}

	testCases := []testCase{
		{
			name: "Create_Server_Success",
			cfg: &config.HTTPConfig{
				Addr:            ":3000",
				ShutdownTimeout: 5,
				ReadTimeout:     10,
				WriteTimeout:    10,
			},
			logger:   logrus.New(),
			checkNil: false,
		},
		{
			name: "Create_Server_With_Nil_Logger",
			cfg: &config.HTTPConfig{
				Addr:            ":3000",
				ShutdownTimeout: 5,
				ReadTimeout:     10,
				WriteTimeout:    10,
			},
			logger:   nil,
			checkNil: false,
		},
		{
			name:     "Create_Server_With_Nil_Config",
			cfg:      nil,
			logger:   logrus.New(),
			checkNil: false,
		},
		{
			name: "Create_Server_With_Empty_Addr",
			cfg: &config.HTTPConfig{
				Addr:            "",
				ShutdownTimeout: 5,
				ReadTimeout:     10,
				WriteTimeout:    10,
			},
			logger:   logrus.New(),
			checkNil: false,
		},
		{
			name: "Create_Server_With_Zero_Timeouts",
			cfg: &config.HTTPConfig{
				Addr:            ":3000",
				ShutdownTimeout: 0,
				ReadTimeout:     0,
				WriteTimeout:    0,
			},
			logger:   logrus.New(),
			checkNil: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			server := NewServer(tc.cfg, mockWebhookService, tc.logger)

			if tc.checkNil {
				if server != nil {
					t.Error("expected server to be nil, got: not nil")
				}
			} else {
				if server == nil {
					t.Error("expected server to be created, got: nil")
					return
				}

				if server.cfg != tc.cfg {
					t.Errorf("expected cfg to be %v, got: %v", tc.cfg, server.cfg)
				}
				if server.logger != tc.logger {
					t.Errorf("expected logger to be %v, got: %v", tc.logger, server.logger)
				}
				if server.webhookService == nil {
					t.Error("expected webhookService to be set, got: nil")
				}
			}
		})
	}
}

func TestServer_Start_Configuration(t *testing.T) {
	type testCase struct {
		name         string
		cfg          *config.HTTPConfig
		checkLogging bool
	}

	testCases := []testCase{
		{
			name: "Start_Server_With_Valid_Config",
			cfg: &config.HTTPConfig{
				Addr:            ":0",
				ShutdownTimeout: 1,
				ReadTimeout:     5,
				WriteTimeout:    5,
			},
			checkLogging: true,
		},
		{
			name: "Start_Server_With_Custom_Timeouts",
			cfg: &config.HTTPConfig{
				Addr:            ":0",
				ShutdownTimeout: 2,
				ReadTimeout:     15,
				WriteTimeout:    20,
			},
			checkLogging: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetOutput(os.Stderr)
			logger.SetLevel(logrus.ErrorLevel)

			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			server := NewServer(tc.cfg, mockWebhookService, logger)

			serverErrors := make(chan error, 1)
			go func() {
				err := server.Start()
				if err != nil {
					serverErrors <- err
				}
			}()

			time.Sleep(100 * time.Millisecond)

			select {
			case err := <-serverErrors:
				if strings.Contains(err.Error(), "failed to start server") {
					t.Errorf("server failed to start: %v", err)
				}
			case <-time.After(200 * time.Millisecond):
			}
		})
	}
}

func TestServer_Start_With_Router(t *testing.T) {
	type testCase struct {
		name         string
		cfg          *config.HTTPConfig
		method       string
		path         string
		requestBody  string
		expectedCode int
		checkBody    bool
		expectedBody string
	}

	testCases := []testCase{
		{
			name: "Start_Server_Health_Endpoint",
			cfg: &config.HTTPConfig{
				Addr:            ":0",
				ShutdownTimeout: 1,
				ReadTimeout:     5,
				WriteTimeout:    5,
			},
			method:       "GET",
			path:         "/health",
			requestBody:  "",
			expectedCode: http.StatusOK,
			checkBody:    true,
			expectedBody: "ok",
		},
		{
			name: "Start_Server_Webhook_Endpoint",
			cfg: &config.HTTPConfig{
				Addr:            ":0",
				ShutdownTimeout: 1,
				ReadTimeout:     5,
				WriteTimeout:    5,
			},
			method:       "POST",
			path:         "/webhook/youtrack",
			requestBody:  `{"test": "data"}`,
			expectedCode: http.StatusOK,
			checkBody:    true,
			expectedBody: "ok",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			server := NewServer(tc.cfg, mockWebhookService, logger)

			router := NewRouter(mockWebhookService, logger)
			testServer := httptest.NewServer(router)
			defer testServer.Close()

			var req *http.Request
			if tc.requestBody != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.requestBody))
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			recorder := httptest.NewRecorder()

			if tc.method == "POST" {
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

			if server.cfg.Addr != tc.cfg.Addr {
				t.Errorf("expected addr %q, got: %q", tc.cfg.Addr, server.cfg.Addr)
			}
		})
	}
}

func TestServer_Start_Error_Handling(t *testing.T) {
	type testCase struct {
		name          string
		cfg           *config.HTTPConfig
		expectedError bool
		checkLogging  bool
	}

	testCases := []testCase{
		{
			name: "Start_Server_With_Invalid_Addr",
			cfg: &config.HTTPConfig{
				Addr:            "invalid:address:format",
				ShutdownTimeout: 1,
				ReadTimeout:     5,
				WriteTimeout:    5,
			},
			expectedError: true,
			checkLogging:  true,
		},
		{
			name: "Start_Server_With_Valid_Config",
			cfg: &config.HTTPConfig{
				Addr:            ":0",
				ShutdownTimeout: 1,
				ReadTimeout:     5,
				WriteTimeout:    5,
			},
			expectedError: false,
			checkLogging:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetOutput(os.Stderr)
			logger.SetLevel(logrus.ErrorLevel)

			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			server := NewServer(tc.cfg, mockWebhookService, logger)

			serverErrors := make(chan error, 1)
			go func() {
				err := server.Start()
				if err != nil {
					serverErrors <- err
				}
			}()

			select {
			case err := <-serverErrors:
				if tc.expectedError {
					if err == nil {
						t.Error("expected error, got: nil")
					}
					if !strings.Contains(err.Error(), "failed to start server") {
						t.Errorf("expected error about failed to start server, got: %v", err)
					}
				} else {
					t.Errorf("unexpected error: %v", err)
				}
			case <-time.After(200 * time.Millisecond):
				if tc.expectedError {
					t.Error("expected error, but server started successfully")
				}
			}
		})
	}
}

func TestServer_Start_Integration(t *testing.T) {
	type testCase struct {
		name         string
		cfg          *config.HTTPConfig
		method       string
		path         string
		requestBody  string
		expectedCode int
	}

	testCases := []testCase{
		{
			name: "Integration_Server_Health_Check",
			cfg: &config.HTTPConfig{
				Addr:            ":0",
				ShutdownTimeout: 1,
				ReadTimeout:     5,
				WriteTimeout:    5,
			},
			method:       "GET",
			path:         "/health",
			requestBody:  "",
			expectedCode: http.StatusOK,
		},
		{
			name: "Integration_Server_Webhook",
			cfg: &config.HTTPConfig{
				Addr:            ":0",
				ShutdownTimeout: 1,
				ReadTimeout:     5,
				WriteTimeout:    5,
			},
			method:       "POST",
			path:         "/webhook/youtrack",
			requestBody:  `{"project": {"name": "Test"}}`,
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
			server := NewServer(tc.cfg, mockWebhookService, logger)

			router := NewRouter(mockWebhookService, logger)
			testServer := httptest.NewServer(router)
			defer testServer.Close()

			var req *http.Request
			if tc.requestBody != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.requestBody))
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			recorder := httptest.NewRecorder()

			if tc.method == "POST" {
				mockWebhookService.EXPECT().ProcessWebhook(gomock.Any()).Return(nil)
			}

			router.ServeHTTP(recorder, req)

			if recorder.Code != tc.expectedCode {
				t.Errorf("expected status code %d, got: %d", tc.expectedCode, recorder.Code)
			}

			if server.cfg == nil {
				t.Error("expected server config to be set")
			}
			if server.webhookService == nil {
				t.Error("expected webhookService to be set")
			}
		})
	}
}

func TestServer_Start_EdgeCases(t *testing.T) {
	type testCase struct {
		name         string
		cfg          *config.HTTPConfig
		logger       *logrus.Logger
		expectedCode int
	}

	testCases := []testCase{
		{
			name: "EdgeCase_Server_With_Nil_Logger",
			cfg: &config.HTTPConfig{
				Addr:            ":0",
				ShutdownTimeout: 1,
				ReadTimeout:     5,
				WriteTimeout:    5,
			},
			logger:       nil,
			expectedCode: http.StatusOK,
		},
		{
			name: "EdgeCase_Server_With_Zero_Shutdown_Timeout",
			cfg: &config.HTTPConfig{
				Addr:            ":0",
				ShutdownTimeout: 0,
				ReadTimeout:     5,
				WriteTimeout:    5,
			},
			logger:       logrus.New(),
			expectedCode: http.StatusOK,
		},
		{
			name: "EdgeCase_Server_With_Large_Timeouts",
			cfg: &config.HTTPConfig{
				Addr:            ":0",
				ShutdownTimeout: 300,
				ReadTimeout:     300,
				WriteTimeout:    300,
			},
			logger:       logrus.New(),
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			server := NewServer(tc.cfg, mockWebhookService, tc.logger)

			if server == nil {
				t.Error("expected server to be created, got: nil")
				return
			}

			if server.cfg.Addr != tc.cfg.Addr {
				t.Errorf("expected addr %q, got: %q", tc.cfg.Addr, server.cfg.Addr)
			}
			if server.cfg.ShutdownTimeout != tc.cfg.ShutdownTimeout {
				t.Errorf("expected shutdown timeout %d, got: %d", tc.cfg.ShutdownTimeout, server.cfg.ShutdownTimeout)
			}
			if server.cfg.ReadTimeout != tc.cfg.ReadTimeout {
				t.Errorf("expected read timeout %d, got: %d", tc.cfg.ReadTimeout, server.cfg.ReadTimeout)
			}
			if server.cfg.WriteTimeout != tc.cfg.WriteTimeout {
				t.Errorf("expected write timeout %d, got: %d", tc.cfg.WriteTimeout, server.cfg.WriteTimeout)
			}

			router := NewRouter(mockWebhookService, tc.logger)
			if router == nil {
				t.Error("expected router to be created, got: nil")
			}
		})
	}
}

func TestServer_Start_Context_Timeout(t *testing.T) {
	type testCase struct {
		name             string
		cfg              *config.HTTPConfig
		shutdownTimeout  time.Duration
		expectedBehavior string
	}

	testCases := []testCase{
		{
			name: "Context_Timeout_Normal_Shutdown",
			cfg: &config.HTTPConfig{
				Addr:            ":0",
				ShutdownTimeout: 1,
				ReadTimeout:     5,
				WriteTimeout:    5,
			},
			shutdownTimeout:  1 * time.Second,
			expectedBehavior: "normal",
		},
		{
			name: "Context_Timeout_Short_Timeout",
			cfg: &config.HTTPConfig{
				Addr:            ":0",
				ShutdownTimeout: 1,
				ReadTimeout:     5,
				WriteTimeout:    5,
			},
			shutdownTimeout:  100 * time.Millisecond,
			expectedBehavior: "normal",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			mockWebhookService := mocks.NewMockWebhookService(ctrl)
			server := NewServer(tc.cfg, mockWebhookService, logger)

			ctx, cancel := context.WithTimeout(context.Background(), tc.shutdownTimeout)
			defer cancel()

			select {
			case <-ctx.Done():
				t.Error("context should not be done immediately")
			default:
			}

			if server.cfg.ShutdownTimeout != tc.cfg.ShutdownTimeout {
				t.Errorf("expected shutdown timeout %d, got: %d", tc.cfg.ShutdownTimeout, server.cfg.ShutdownTimeout)
			}
		})
	}
}
