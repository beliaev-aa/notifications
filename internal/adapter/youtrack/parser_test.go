package youtrack

import (
	"github.com/beliaev-aa/notifications/internal/domain/port"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"github.com/beliaev-aa/notifications/tests/mocks"
	"github.com/golang/mock/gomock"
	"strings"
	"testing"
)

func TestNewParser(t *testing.T) {
	type testCase struct {
		name           string
		checkInterface bool
		checkNil       bool
	}

	testCases := []testCase{
		{
			name:           "Create_Parser_Success",
			checkInterface: true,
			checkNil:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProjectConfig := mocks.NewMockProjectConfigService(ctrl)
			p := NewParser(mockProjectConfig)

			if tc.checkNil {
				if p != nil {
					t.Error("expected parser to be nil, got: not nil")
				}
			} else {
				if p == nil {
					t.Error("expected parser to be created, got: nil")
					return
				}

				if tc.checkInterface {
					if _, ok := p.(parser.YoutrackParser); !ok {
						t.Error("expected parser to implement YoutrackParser interface")
					}
				}
			}
		})
	}
}

func TestParser_ParseJSON(t *testing.T) {
	type testCase struct {
		name          string
		payload       []byte
		expectedError bool
		expectedNil   bool
		validateFunc  func(*testing.T, *parser.YoutrackWebhookPayload)
	}

	projectName := "Test Project"
	issueSummary := "Test Issue"
	issueURL := "https://youtrack.test.com/issue/123"

	validJSON := `{
		"project": {
			"name": "Test Project",
			"presentation": "Test Project Presentation"
		},
		"issue": {
			"summary": "Test Issue",
			"url": "https://youtrack.test.com/issue/123",
			"state": {
				"name": "In Progress"
			},
			"priority": {
				"name": "High"
			},
			"assignee": {
				"fullName": "Jane Doe",
				"login": "login"
			}
		},
		"updater": {
			"fullName": "John Doe",
			"login": "login",
			"email": "john@example.com"
		},
		"changes": [
			{
				"field": "State",
				"oldValue": "To Do",
				"newValue": "In Progress"
			}
		]
	}`

	testCases := []testCase{
		{
			name:          "Parse_Valid_JSON",
			payload:       []byte(validJSON),
			expectedError: false,
			expectedNil:   false,
			validateFunc: func(t *testing.T, payload *parser.YoutrackWebhookPayload) {
				if payload == nil {
					t.Error("expected payload to be not nil")
					return
				}

				if payload.Project == nil {
					t.Error("expected project to be not nil")
				} else {
					if payload.Project.Name == nil || *payload.Project.Name != projectName {
						t.Errorf("expected project name %q, got: %v", projectName, payload.Project.Name)
					}
				}

				if payload.Issue.Summary != issueSummary {
					t.Errorf("expected issue summary %q, got: %q", issueSummary, payload.Issue.Summary)
				}

				if payload.Issue.URL != issueURL {
					t.Errorf("expected issue URL %q, got: %q", issueURL, payload.Issue.URL)
				}

				if len(payload.Changes) != 1 {
					t.Errorf("expected 1 change, got: %d", len(payload.Changes))
				} else if payload.Changes[0].Field != "State" {
					t.Errorf("expected change field %q, got: %q", "State", payload.Changes[0].Field)
				}
			},
		},
		{
			name:          "Parse_Invalid_JSON",
			payload:       []byte(`{invalid json}`),
			expectedError: true,
			expectedNil:   true,
			validateFunc:  nil,
		},
		{
			name:          "Parse_Empty_JSON",
			payload:       []byte(`{}`),
			expectedError: false,
			expectedNil:   false,
			validateFunc: func(t *testing.T, payload *parser.YoutrackWebhookPayload) {
				if payload == nil {
					t.Error("expected payload to be not nil")
				}
			},
		},
		{
			name:          "Parse_Empty_Bytes",
			payload:       []byte(``),
			expectedError: true,
			expectedNil:   true,
			validateFunc:  nil,
		},
		{
			name:          "Parse_Nil_Bytes",
			payload:       nil,
			expectedError: true,
			expectedNil:   true,
			validateFunc:  nil,
		},
		{
			name:          "Parse_Partial_JSON",
			payload:       []byte(`{"issue": {"summary": "Test"}}`),
			expectedError: false,
			expectedNil:   false,
			validateFunc: func(t *testing.T, payload *parser.YoutrackWebhookPayload) {
				if payload == nil {
					t.Error("expected payload to be not nil")
					return
				}
				if payload.Issue.Summary != "Test" {
					t.Errorf("expected issue summary %q, got: %q", "Test", payload.Issue.Summary)
				}
			},
		},
		{
			name:          "Parse_JSON_With_Null_Values",
			payload:       []byte(`{"project": null, "issue": {"summary": "Test"}}`),
			expectedError: false,
			expectedNil:   false,
			validateFunc: func(t *testing.T, payload *parser.YoutrackWebhookPayload) {
				if payload == nil {
					t.Error("expected payload to be not nil")
					return
				}
				if payload.Project != nil {
					t.Error("expected project to be nil")
				}
				if payload.Issue.Summary != "Test" {
					t.Errorf("expected issue summary %q, got: %q", "Test", payload.Issue.Summary)
				}
			},
		},
		{
			name:          "Parse_JSON_With_Array_Changes",
			payload:       []byte(`{"changes": [{"field": "State", "oldValue": "A", "newValue": "B"}, {"field": "Priority", "oldValue": "Low", "newValue": "High"}]}`),
			expectedError: false,
			expectedNil:   false,
			validateFunc: func(t *testing.T, payload *parser.YoutrackWebhookPayload) {
				if payload == nil {
					t.Error("expected payload to be not nil")
					return
				}
				if len(payload.Changes) != 2 {
					t.Errorf("expected 2 changes, got: %d", len(payload.Changes))
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProjectConfig := mocks.NewMockProjectConfigService(ctrl)
			p := NewParser(mockProjectConfig)

			result, err := p.ParseJSON(tc.payload)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error, got: nil")
				} else {
					if err.Error() == "" {
						t.Error("expected error message, got: empty string")
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if tc.expectedNil {
				if result != nil {
					t.Error("expected result to be nil, got: not nil")
				}
			} else {
				if result == nil {
					t.Error("expected result to be not nil, got: nil")
				}
			}

			if tc.validateFunc != nil && result != nil {
				tc.validateFunc(t, result)
			}
		})
	}
}

func TestParser_NewFormatter(t *testing.T) {
	type testCase struct {
		name           string
		checkInterface bool
		checkNil       bool
	}

	testCases := []testCase{
		{
			name:           "NewFormatter_Success",
			checkInterface: true,
			checkNil:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProjectConfig := mocks.NewMockProjectConfigService(ctrl)
			p := NewParser(mockProjectConfig)

			formatter := p.NewFormatter()

			if tc.checkNil {
				if formatter != nil {
					t.Error("expected formatter to be nil, got: not nil")
				}
			} else {
				if formatter == nil {
					t.Error("expected formatter to be created, got: nil")
					return
				}

				if tc.checkInterface {
					emptyPayload := &parser.YoutrackWebhookPayload{}
					_ = formatter.Format(emptyPayload, "test")
				}
			}
		})
	}
}

func TestParser_Integration(t *testing.T) {
	type testCase struct {
		name           string
		payload        []byte
		checkFormatter bool
	}

	validJSON := `{
		"project": {
			"name": "Test Project"
		},
		"issue": {
			"summary": "Test Issue",
			"url": "https://test.com"
		}
	}`

	testCases := []testCase{
		{
			name:           "Full_Integration_Test",
			payload:        []byte(validJSON),
			checkFormatter: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProjectConfig := mocks.NewMockProjectConfigService(ctrl)
			youtrackParser := NewParser(mockProjectConfig)

			payload, err := youtrackParser.ParseJSON(tc.payload)
			if err != nil {
				t.Fatalf("unexpected error parsing JSON: %v", err)
			}

			if payload == nil {
				t.Fatal("expected payload to be not nil")
			}

			if tc.checkFormatter {
				formatter := youtrackParser.NewFormatter()
				if formatter == nil {
					t.Fatal("expected formatter to be created, got: nil")
				}

				formatted := formatter.Format(payload, "test")
				if formatted == "" {
					t.Error("expected formatted message to be not empty")
				}
			}
		})
	}
}

func TestParser_ParseJSON_EdgeCases(t *testing.T) {
	type testCase struct {
		name          string
		payload       []byte
		expectedError bool
	}

	testCases := []testCase{
		{
			name:          "Parse_JSON_With_Only_Whitespace",
			payload:       []byte(`   `),
			expectedError: true,
		},
		{
			name:          "Parse_JSON_With_Unicode",
			payload:       []byte(`{"issue": {"summary": "Тест 🚀"}}`),
			expectedError: false,
		},
		{
			name:          "Parse_JSON_With_Special_Characters",
			payload:       []byte(`{"issue": {"summary": "Test & <test> \"quotes\""}}`),
			expectedError: false,
		},
		{
			name:          "Parse_JSON_With_Number_Fields",
			payload:       []byte(`{"issue": {"summary": "Test", "id": 123}}`),
			expectedError: false,
		},
		{
			name:          "Parse_JSON_With_Boolean_Fields",
			payload:       []byte(`{"issue": {"summary": "Test", "resolved": true}}`),
			expectedError: false,
		},
		{
			name:          "Parse_Malformed_JSON_Missing_Quote",
			payload:       []byte(`{"issue": {summary: "Test"}}`),
			expectedError: true,
		},
		{
			name:          "Parse_Malformed_JSON_Missing_Comma",
			payload:       []byte(`{"issue": {"summary": "Test" "url": "test.com"}}`),
			expectedError: true,
		},
		{
			name:          "Parse_Malformed_JSON_Unclosed_Brace",
			payload:       []byte(`{"issue": {"summary": "Test"}`),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProjectConfig := mocks.NewMockProjectConfigService(ctrl)
			p := NewParser(mockProjectConfig)

			result, err := p.ParseJSON(tc.payload)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error, got: nil")
				}
				if result != nil {
					t.Error("expected result to be nil on error, got: not nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Error("expected result to be not nil, got: nil")
				}
			}
		})
	}
}

func TestParser_GetAllowedChannels(t *testing.T) {
	type testCase struct {
		name             string
		payload          *parser.YoutrackWebhookPayload
		projectAllowed   bool
		allowedChannels  []string
		expectedChannels []string
	}

	projectName := "TestProject"
	emptyProjectName := ""

	testCases := []testCase{
		{
			name: "GetAllowedChannels_Project_Allowed",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{Name: &projectName},
			},
			projectAllowed:   true,
			allowedChannels:  []string{port.ChannelLogger, port.ChannelTelegram},
			expectedChannels: []string{port.ChannelLogger, port.ChannelTelegram},
		},
		{
			name: "GetAllowedChannels_Project_Not_Allowed",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{Name: &projectName},
			},
			projectAllowed:   false,
			allowedChannels:  []string{},
			expectedChannels: []string{},
		},
		{
			name: "GetAllowedChannels_Project_Nil",
			payload: &parser.YoutrackWebhookPayload{
				Project: nil,
			},
			projectAllowed:   false,
			allowedChannels:  []string{},
			expectedChannels: []string{},
		},
		{
			name: "GetAllowedChannels_Project_Name_Nil",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{Name: nil},
			},
			projectAllowed:   false,
			allowedChannels:  []string{},
			expectedChannels: []string{},
		},
		{
			name: "GetAllowedChannels_Project_Name_Empty",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{Name: &emptyProjectName},
			},
			projectAllowed:   false,
			allowedChannels:  []string{},
			expectedChannels: []string{},
		},
		{
			name: "GetAllowedChannels_Only_Logger",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{Name: &projectName},
			},
			projectAllowed:   true,
			allowedChannels:  []string{port.ChannelLogger},
			expectedChannels: []string{port.ChannelLogger},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProjectConfig := mocks.NewMockProjectConfigService(ctrl)
			p := NewParser(mockProjectConfig)

			if tc.payload.Project != nil && tc.payload.Project.Name != nil && *tc.payload.Project.Name != "" {
				normalizedName := strings.ToLower(*tc.payload.Project.Name)
				mockProjectConfig.EXPECT().IsProjectAllowed(normalizedName).Return(tc.projectAllowed)
				if tc.projectAllowed {
					mockProjectConfig.EXPECT().GetAllowedChannels(normalizedName).Return(tc.allowedChannels)
				}
			}

			channels := p.GetAllowedChannels(tc.payload)

			if len(channels) != len(tc.expectedChannels) {
				t.Errorf("expected %d channels, got: %d", len(tc.expectedChannels), len(channels))
			}

			for i, expected := range tc.expectedChannels {
				if i >= len(channels) || channels[i] != expected {
					t.Errorf("expected channel %q at index %d, got: %v", expected, i, channels)
					break
				}
			}
		})
	}
}

func TestParser_GetTelegramChatID(t *testing.T) {
	type testCase struct {
		name              string
		projectName       string
		chatID            string
		hasChatID         bool
		expectedChatID    string
		expectedHasChatID bool
	}

	testCases := []testCase{
		{
			name:              "GetTelegramChatID_Project_With_Telegram",
			projectName:       "TestProject",
			chatID:            "test_chat_id",
			hasChatID:         true,
			expectedChatID:    "test_chat_id",
			expectedHasChatID: true,
		},
		{
			name:              "GetTelegramChatID_Project_Without_Telegram",
			projectName:       "TestProject",
			chatID:            "",
			hasChatID:         false,
			expectedChatID:    "",
			expectedHasChatID: false,
		},
		{
			name:              "GetTelegramChatID_Non_Existent_Project",
			projectName:       "NonExistentProject",
			chatID:            "",
			hasChatID:         false,
			expectedChatID:    "",
			expectedHasChatID: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProjectConfig := mocks.NewMockProjectConfigService(ctrl)
			normalizedName := strings.ToLower(tc.projectName)
			mockProjectConfig.EXPECT().GetTelegramChatID(normalizedName).Return(tc.chatID, tc.hasChatID)

			p := NewParser(mockProjectConfig)

			chatID, hasChatID := p.GetTelegramChatID(tc.projectName)

			if chatID != tc.expectedChatID {
				t.Errorf("expected chatID %q, got: %q", tc.expectedChatID, chatID)
			}

			if hasChatID != tc.expectedHasChatID {
				t.Errorf("expected hasChatID %v, got: %v", tc.expectedHasChatID, hasChatID)
			}
		})
	}
}

func TestParser_GetAllowedChannels_CaseInsensitive(t *testing.T) {
	type testCase struct {
		name             string
		projectName      string
		projectNameCase  string
		allowedChannels  []string
		expectedChannels []string
	}

	testCases := []testCase{
		{
			name:             "GetAllowedChannels_Uppercase_Project_Name",
			projectName:      "project1",
			projectNameCase:  "PROJECT1",
			allowedChannels:  []string{"logger", "telegram"},
			expectedChannels: []string{"logger", "telegram"},
		},
		{
			name:             "GetAllowedChannels_Mixed_Case_Project_Name",
			projectName:      "project1",
			projectNameCase:  "Project1",
			allowedChannels:  []string{"logger"},
			expectedChannels: []string{"logger"},
		},
		{
			name:             "GetAllowedChannels_Lowercase_Project_Name",
			projectName:      "PROJECT1",
			projectNameCase:  "project1",
			allowedChannels:  []string{"telegram"},
			expectedChannels: []string{"telegram"},
		},
		{
			name:             "GetAllowedChannels_Mixed_Case_With_Underscore",
			projectName:      "project_test",
			projectNameCase:  "Project_Test",
			allowedChannels:  []string{"logger", "telegram"},
			expectedChannels: []string{"logger", "telegram"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProjectConfig := mocks.NewMockProjectConfigService(ctrl)
			normalizedName := strings.ToLower(tc.projectNameCase)
			mockProjectConfig.EXPECT().IsProjectAllowed(normalizedName).Return(true)
			mockProjectConfig.EXPECT().GetAllowedChannels(normalizedName).Return(tc.allowedChannels)

			p := NewParser(mockProjectConfig)

			payload := &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{
					Name: &tc.projectNameCase,
				},
			}

			channels := p.GetAllowedChannels(payload)

			if len(channels) != len(tc.expectedChannels) {
				t.Errorf("expected %d channels, got: %d", len(tc.expectedChannels), len(channels))
			}

			for i, expectedChannel := range tc.expectedChannels {
				if i >= len(channels) {
					t.Errorf("expected channel %q at index %d, got: no channel", expectedChannel, i)
					continue
				}
				if channels[i] != expectedChannel {
					t.Errorf("expected channel %q at index %d, got: %q", expectedChannel, i, channels[i])
				}
			}
		})
	}
}

func TestParser_GetTelegramChatID_CaseInsensitive(t *testing.T) {
	type testCase struct {
		name              string
		projectName       string
		projectNameCase   string
		chatID            string
		hasChatID         bool
		expectedChatID    string
		expectedHasChatID bool
	}

	testCases := []testCase{
		{
			name:              "GetTelegramChatID_Uppercase_Project_Name",
			projectName:       "project1",
			projectNameCase:   "PROJECT1",
			chatID:            "123456789",
			hasChatID:         true,
			expectedChatID:    "123456789",
			expectedHasChatID: true,
		},
		{
			name:              "GetTelegramChatID_Mixed_Case_Project_Name",
			projectName:       "project1",
			projectNameCase:   "Project1",
			chatID:            "987654321",
			hasChatID:         true,
			expectedChatID:    "987654321",
			expectedHasChatID: true,
		},
		{
			name:              "GetTelegramChatID_Lowercase_Project_Name",
			projectName:       "PROJECT1",
			projectNameCase:   "project1",
			chatID:            "111222333",
			hasChatID:         true,
			expectedChatID:    "111222333",
			expectedHasChatID: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProjectConfig := mocks.NewMockProjectConfigService(ctrl)

			normalizedName := strings.ToLower(tc.projectNameCase)
			mockProjectConfig.EXPECT().GetTelegramChatID(normalizedName).Return(tc.chatID, tc.hasChatID)

			p := NewParser(mockProjectConfig)

			chatID, hasChatID := p.GetTelegramChatID(tc.projectNameCase)

			if chatID != tc.expectedChatID {
				t.Errorf("expected chatID %q, got: %q", tc.expectedChatID, chatID)
			}

			if hasChatID != tc.expectedHasChatID {
				t.Errorf("expected hasChatID %v, got: %v", tc.expectedHasChatID, hasChatID)
			}
		})
	}
}

func TestParser_GetVKTeamsChatID(t *testing.T) {
	type testCase struct {
		name              string
		projectName       string
		chatID            string
		hasChatID         bool
		expectedChatID    string
		expectedHasChatID bool
	}

	testCases := []testCase{
		{
			name:              "GetVKTeamsChatID_Project_With_VKTeams",
			projectName:       "TestProject",
			chatID:            "test_chat_id",
			hasChatID:         true,
			expectedChatID:    "test_chat_id",
			expectedHasChatID: true,
		},
		{
			name:              "GetVKTeamsChatID_Project_Without_VKTeams",
			projectName:       "TestProject",
			chatID:            "",
			hasChatID:         false,
			expectedChatID:    "",
			expectedHasChatID: false,
		},
		{
			name:              "GetVKTeamsChatID_Non_Existent_Project",
			projectName:       "NonExistentProject",
			chatID:            "",
			hasChatID:         false,
			expectedChatID:    "",
			expectedHasChatID: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProjectConfig := mocks.NewMockProjectConfigService(ctrl)
			normalizedName := strings.ToLower(tc.projectName)
			mockProjectConfig.EXPECT().GetVKTeamsChatID(normalizedName).Return(tc.chatID, tc.hasChatID)

			p := NewParser(mockProjectConfig)

			chatID, hasChatID := p.GetVKTeamsChatID(tc.projectName)

			if chatID != tc.expectedChatID {
				t.Errorf("expected chatID %q, got: %q", tc.expectedChatID, chatID)
			}

			if hasChatID != tc.expectedHasChatID {
				t.Errorf("expected hasChatID %v, got: %v", tc.expectedHasChatID, hasChatID)
			}
		})
	}
}

func TestParser_GetVKTeamsChatID_CaseInsensitive(t *testing.T) {
	type testCase struct {
		name              string
		projectName       string
		projectNameCase   string
		chatID            string
		hasChatID         bool
		expectedChatID    string
		expectedHasChatID bool
	}

	testCases := []testCase{
		{
			name:              "GetVKTeamsChatID_Uppercase_Project_Name",
			projectName:       "project1",
			projectNameCase:   "PROJECT1",
			chatID:            "123456789",
			hasChatID:         true,
			expectedChatID:    "123456789",
			expectedHasChatID: true,
		},
		{
			name:              "GetVKTeamsChatID_Mixed_Case_Project_Name",
			projectName:       "project1",
			projectNameCase:   "Project1",
			chatID:            "987654321",
			hasChatID:         true,
			expectedChatID:    "987654321",
			expectedHasChatID: true,
		},
		{
			name:              "GetVKTeamsChatID_Lowercase_Project_Name",
			projectName:       "PROJECT1",
			projectNameCase:   "project1",
			chatID:            "111222333",
			hasChatID:         true,
			expectedChatID:    "111222333",
			expectedHasChatID: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProjectConfig := mocks.NewMockProjectConfigService(ctrl)

			normalizedName := strings.ToLower(tc.projectNameCase)
			mockProjectConfig.EXPECT().GetVKTeamsChatID(normalizedName).Return(tc.chatID, tc.hasChatID)

			p := NewParser(mockProjectConfig)

			chatID, hasChatID := p.GetVKTeamsChatID(tc.projectNameCase)

			if chatID != tc.expectedChatID {
				t.Errorf("expected chatID %q, got: %q", tc.expectedChatID, chatID)
			}

			if hasChatID != tc.expectedHasChatID {
				t.Errorf("expected hasChatID %v, got: %v", tc.expectedHasChatID, hasChatID)
			}
		})
	}
}
