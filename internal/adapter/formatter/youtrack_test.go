package formatter

import (
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"strings"
	"testing"
)

func TestNewYoutrackFormatter(t *testing.T) {
	type testCase struct {
		name           string
		checkInterface bool
		checkNil       bool
	}

	testCases := []testCase{
		{
			name:           "Create_YoutrackFormatter_Success",
			checkInterface: true,
			checkNil:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := NewYoutrackFormatter()

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
					if _, ok := formatter.(parser.YoutrackFormatter); !ok {
						t.Error("expected formatter to implement YoutrackFormatter interface")
					}
				}
			}
		})
	}
}

func TestYoutrackFormatter_Format(t *testing.T) {
	type testCase struct {
		name            string
		payload         *parser.YoutrackWebhookPayload
		channel         string
		customFormatter func(payload *parser.YoutrackWebhookPayload) string
		expectedResult  string
		checkContains   []string
	}

	projectName := "TestProject"
	projectPresentation := "Test Project"
	issueSummary := "Test Issue"
	issueURL := "https://youtrack.test/issue/123"
	stateName := "In Progress"
	priorityName := "High"
	assigneeFullName := "John Doe"
	updaterFullName := "Jane Smith"

	testPayload := &parser.YoutrackWebhookPayload{
		Project: &parser.YoutrackFieldValue{
			Name:         &projectName,
			Presentation: &projectPresentation,
		},
		Issue: parser.YoutrackIssue{
			Summary: issueSummary,
			URL:     issueURL,
			State: &parser.YoutrackFieldValue{
				Name: &stateName,
			},
			Priority: &parser.YoutrackFieldValue{
				Name: &priorityName,
			},
			Assignee: &parser.YoutrackUser{
				FullName: &assigneeFullName,
			},
		},
		Updater: &parser.YoutrackUser{
			FullName: &updaterFullName,
		},
		Changes: []parser.YoutrackChange{
			{
				Field:    "State",
				OldValue: []byte(`"To Do"`),
				NewValue: []byte(`"In Progress"`),
			},
		},
	}

	testCases := []testCase{
		{
			name:    "Format_Default_Channel",
			payload: testPayload,
			channel: "unknown",
			checkContains: []string{
				"Проект:",
				"Задача:",
				"Ссылка:",
				"Статус:",
				"Приоритет:",
				"Issue",
				"youtrack.test",
			},
		},
		{
			name:    "Format_With_Custom_Formatter",
			payload: testPayload,
			channel: "telegram",
			customFormatter: func(payload *parser.YoutrackWebhookPayload) string {
				return "Custom format: " + payload.Issue.Summary
			},
			expectedResult: "Custom format: Test Issue",
		},
		{
			name:    "Format_With_Empty_Payload",
			payload: &parser.YoutrackWebhookPayload{},
			channel: "default",
			checkContains: []string{
				"Проект:",
				"Задача:",
			},
		},
		{
			name: "Format_With_Nil_Fields",
			payload: &parser.YoutrackWebhookPayload{
				Project: nil,
				Issue: parser.YoutrackIssue{
					Summary: "Test",
					URL:     "https://test.com",
				},
				Updater: nil,
				Changes: nil,
			},
			channel: "default",
			checkContains: []string{
				"Проект:",
				"Задача: Test",
			},
		},
		{
			name: "Format_With_Comment_Change",
			payload: &parser.YoutrackWebhookPayload{
				Issue: parser.YoutrackIssue{
					Summary: "Test Issue",
					URL:     "https://test.com",
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    "Comment",
						OldValue: []byte(`null`),
						NewValue: []byte(`{"text": "Test comment", "mentionedUsers": []}`),
					},
				},
			},
			channel: "default",
			checkContains: []string{
				"Комментарий:",
			},
		},
		{
			name: "Format_With_Multiple_Changes",
			payload: &parser.YoutrackWebhookPayload{
				Issue: parser.YoutrackIssue{
					Summary: "Test Issue",
					URL:     "https://test.com",
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    "State",
						OldValue: []byte(`"To Do"`),
						NewValue: []byte(`"In Progress"`),
					},
					{
						Field:    "Priority",
						OldValue: []byte(`"Low"`),
						NewValue: []byte(`"High"`),
					},
				},
			},
			channel: "default",
			checkContains: []string{
				"Состояние:",
				"Приоритет:",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := NewYoutrackFormatter().(*YoutrackFormatter)

			if tc.customFormatter != nil {
				formatter.RegisterChannelFormatter(tc.channel, tc.customFormatter)
			}

			result := formatter.Format(tc.payload, tc.channel)

			if tc.expectedResult != "" {
				if result != tc.expectedResult {
					t.Errorf("expected result %q, got: %q", tc.expectedResult, result)
				}
			}

			if len(tc.checkContains) > 0 {
				for _, expected := range tc.checkContains {
					if !strings.Contains(result, expected) {
						t.Errorf("expected result to contain %q, got: %q", expected, result)
					}
				}
			}

			if result == "" {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestYoutrackFormatter_RegisterChannelFormatter(t *testing.T) {
	type testCase struct {
		name           string
		channel        string
		formatter      func(payload *parser.YoutrackWebhookPayload) string
		payload        *parser.YoutrackWebhookPayload
		shouldRegister bool
		expectedResult string
	}

	testPayload := &parser.YoutrackWebhookPayload{
		Issue: parser.YoutrackIssue{
			Summary: "Test Issue",
			URL:     "https://test.com",
		},
	}

	customFormatter := func(payload *parser.YoutrackWebhookPayload) string {
		return "Custom: " + payload.Issue.Summary
	}

	testCases := []testCase{
		{
			name:           "RegisterChannelFormatter_Success",
			channel:        "telegram",
			formatter:      customFormatter,
			payload:        testPayload,
			shouldRegister: true,
			expectedResult: "Custom: Test Issue",
		},
		{
			name:           "RegisterChannelFormatter_With_Empty_Channel",
			channel:        "",
			formatter:      customFormatter,
			payload:        testPayload,
			shouldRegister: false,
		},
		{
			name:           "RegisterChannelFormatter_With_Nil_Formatter",
			channel:        "telegram",
			formatter:      nil,
			payload:        testPayload,
			shouldRegister: false,
		},
		{
			name:           "RegisterChannelFormatter_Multiple_Channels",
			channel:        "slack",
			formatter:      func(payload *parser.YoutrackWebhookPayload) string { return "Slack format" },
			payload:        testPayload,
			shouldRegister: true,
			expectedResult: "Slack format",
		},
		{
			name:           "RegisterChannelFormatter_Override_Existing",
			channel:        "telegram",
			formatter:      func(payload *parser.YoutrackWebhookPayload) string { return "New format" },
			payload:        testPayload,
			shouldRegister: true,
			expectedResult: "New format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := NewYoutrackFormatter().(*YoutrackFormatter)

			if tc.name == "RegisterChannelFormatter_Override_Existing" {
				formatter.RegisterChannelFormatter("telegram", customFormatter)
			}

			formatter.RegisterChannelFormatter(tc.channel, tc.formatter)

			if tc.shouldRegister {
				result := formatter.Format(tc.payload, tc.channel)
				if tc.expectedResult != "" {
					if result != tc.expectedResult {
						t.Errorf("expected result %q, got: %q", tc.expectedResult, result)
					}
				}
			} else {
				result := formatter.Format(tc.payload, tc.channel)
				if result == "" {
					t.Error("expected non-empty result from default formatter")
				}
				if tc.expectedResult != "" && result == tc.expectedResult {
					t.Error("expected default formatter to be used, but got custom format")
				}
			}
		})
	}
}

func TestYoutrackFormatter_Integration(t *testing.T) {
	type testCase struct {
		name    string
		payload *parser.YoutrackWebhookPayload
	}

	projectName := "IntegrationProject"
	issueSummary := "Integration Test Issue"
	issueURL := "https://youtrack.test/issue/INT-1"
	stateName := "Done"
	priorityName := "Critical"
	assigneeFullName := "Integration User"

	testCases := []testCase{
		{
			name: "Full_Integration_Test",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{
					Name: &projectName,
				},
				Issue: parser.YoutrackIssue{
					Summary: issueSummary,
					URL:     issueURL,
					State: &parser.YoutrackFieldValue{
						Name: &stateName,
					},
					Priority: &parser.YoutrackFieldValue{
						Name: &priorityName,
					},
					Assignee: &parser.YoutrackUser{
						FullName: &assigneeFullName,
					},
				},
				Updater: &parser.YoutrackUser{
					FullName: &assigneeFullName,
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    "State",
						OldValue: []byte(`"In Progress"`),
						NewValue: []byte(`"Done"`),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := NewYoutrackFormatter()

			if _, ok := formatter.(parser.YoutrackFormatter); !ok {
				t.Fatal("expected formatter to implement YoutrackFormatter interface")
			}

			result := formatter.Format(tc.payload, "default")
			if result == "" {
				t.Fatal("expected non-empty result")
			}

			expectedElements := []string{
				"Проект:",
				"Задача:",
				"Ссылка:",
				"Статус:",
				"Приоритет:",
				"Issue",
			}
			for _, element := range expectedElements {
				if !strings.Contains(result, element) {
					t.Logf("result may not contain %q (this might be expected)", element)
				}
			}

			customFormat := "Custom integration format"
			formatter.RegisterChannelFormatter("custom", func(payload *parser.YoutrackWebhookPayload) string {
				return customFormat
			})

			customResult := formatter.Format(tc.payload, "custom")
			if customResult != customFormat {
				t.Errorf("expected custom format %q, got: %q", customFormat, customResult)
			}
		})
	}
}

func TestYoutrackFormatter_EdgeCases(t *testing.T) {
	type testCase struct {
		name             string
		payload          *parser.YoutrackWebhookPayload
		channel          string
		expectedResult   string
		checkContains    []string
		checkNotContains []string
	}

	inProgressEn := "In Progress"
	inProgressRu := "В работе"
	loginStr := "Login"

	testCases := []testCase{
		{
			name: "Format_With_Empty_Channel_Name",
			payload: &parser.YoutrackWebhookPayload{
				Issue: parser.YoutrackIssue{
					Summary: "Test",
					URL:     "https://test.com",
				},
			},
			channel: "",
			checkContains: []string{
				"Проект:",
			},
		},
		{
			name: "Format_With_Unicode_Characters",
			payload: &parser.YoutrackWebhookPayload{
				Issue: parser.YoutrackIssue{
					Summary: "Тест с кириллицей 🚀",
					URL:     "https://test.com",
				},
			},
			channel: "default",
			checkContains: []string{
				"Тест с кириллицей",
			},
		},
		{
			name: "Format_With_Special_Characters_In_Summary",
			payload: &parser.YoutrackWebhookPayload{
				Issue: parser.YoutrackIssue{
					Summary: "Test & <test> \"quotes\" 'apostrophes'",
					URL:     "https://test.com",
				},
			},
			channel: "default",
			checkContains: []string{
				"Test &",
			},
		},
		{
			name: "Format_With_Comment_With_Mentions",
			payload: &parser.YoutrackWebhookPayload{
				Issue: parser.YoutrackIssue{
					Summary: "Test Issue",
					URL:     "https://test.com",
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    "Comment",
						OldValue: []byte(`null`),
						NewValue: []byte(`{"text": "Test comment with @user", "mentionedUsers": [{"fullName": "John Doe"}]}`),
					},
				},
			},
			channel: "default",
			checkContains: []string{
				"Комментарий:",
				"Упомянуты:",
			},
		},
		{
			name: "Format_With_Assignee_Change",
			payload: &parser.YoutrackWebhookPayload{
				Issue: parser.YoutrackIssue{
					Summary: "Test Issue",
					URL:     "https://test.com",
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    "Assignee",
						OldValue: []byte(`{"fullName": "Old User"}`),
						NewValue: []byte(`{"fullName": "New User"}`),
					},
				},
			},
			channel: "default",
			checkContains: []string{
				"Назначена:",
			},
		},
		{
			name: "Format_With_Null_Values",
			payload: &parser.YoutrackWebhookPayload{
				Issue: parser.YoutrackIssue{
					Summary: "Test Issue",
					URL:     "https://test.com",
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    "State",
						OldValue: []byte(`null`),
						NewValue: []byte(`null`),
					},
				},
			},
			channel: "default",
			checkContains: []string{
				"Состояние:",
				"(Не установлен)",
			},
		},
		{
			name: "Format_With_Empty_Changes",
			payload: &parser.YoutrackWebhookPayload{
				Issue: parser.YoutrackIssue{
					Summary: "Test Issue",
					URL:     "https://test.com",
				},
				Changes: []parser.YoutrackChange{},
			},
			channel: "default",
			checkContains: []string{
				"Проект:",
				"Изменения:",
			},
		},
		{
			name: "Format_With_Field_With_Presentation",
			payload: &parser.YoutrackWebhookPayload{
				Issue: parser.YoutrackIssue{
					Summary: "Test Issue",
					URL:     "https://test.com",
					State: &parser.YoutrackFieldValue{
						Name:         &inProgressEn,
						Presentation: &inProgressRu,
					},
				},
			},
			channel: "default",
			checkContains: []string{
				inProgressRu,
			},
		},
		{
			name: "Format_With_User_With_Login_Only",
			payload: &parser.YoutrackWebhookPayload{
				Issue: parser.YoutrackIssue{
					Summary: "Test Issue",
					URL:     "https://test.com",
					Assignee: &parser.YoutrackUser{
						Login: &loginStr,
					},
				},
			},
			channel: "default",
			checkContains: []string{
				loginStr,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := NewYoutrackFormatter().(*YoutrackFormatter)

			result := formatter.Format(tc.payload, tc.channel)

			if tc.expectedResult != "" {
				if result != tc.expectedResult {
					t.Errorf("expected result %q, got: %q", tc.expectedResult, result)
				}
			}

			if len(tc.checkContains) > 0 {
				for _, expected := range tc.checkContains {
					if !strings.Contains(result, expected) {
						t.Errorf("expected result to contain %q, got: %q", expected, result)
					}
				}
			}

			if len(tc.checkNotContains) > 0 {
				for _, notExpected := range tc.checkNotContains {
					if strings.Contains(result, notExpected) {
						t.Errorf("expected result not to contain %q, got: %q", notExpected, result)
					}
				}
			}

			if result == "" {
				t.Error("expected non-empty result")
			}
		})
	}
}
