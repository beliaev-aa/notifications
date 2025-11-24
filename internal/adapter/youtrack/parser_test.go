package youtrack

import (
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
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
			p := NewParser()

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
				"login": "janedoe"
			}
		},
		"updater": {
			"fullName": "John Doe",
			"login": "johndoe",
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
			p := NewParser()

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
			p := NewParser()

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
			parser := NewParser()

			payload, err := parser.ParseJSON(tc.payload)
			if err != nil {
				t.Fatalf("unexpected error parsing JSON: %v", err)
			}

			if payload == nil {
				t.Fatal("expected payload to be not nil")
			}

			if tc.checkFormatter {
				formatter := parser.NewFormatter()
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
			p := NewParser()

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
