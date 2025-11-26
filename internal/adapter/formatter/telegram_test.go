package formatter

import (
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"strings"
	"testing"
)

func TestEscapeMarkdownV2(t *testing.T) {
	type testCase struct {
		name           string
		text           string
		expectedResult string
	}

	testCases := []testCase{
		{
			name:           "Escape_Empty_String",
			text:           "",
			expectedResult: "",
		},
		{
			name:           "Escape_Simple_Text",
			text:           "Simple text",
			expectedResult: "Simple text",
		},
		{
			name:           "Escape_Underscore",
			text:           "text_with_underscore",
			expectedResult: "text\\_with\\_underscore",
		},
		{
			name:           "Escape_Asterisk",
			text:           "text *with* asterisk",
			expectedResult: "text \\*with\\* asterisk",
		},
		{
			name:           "Escape_Brackets",
			text:           "text [with] brackets",
			expectedResult: "text \\[with\\] brackets",
		},
		{
			name:           "Escape_Parentheses",
			text:           "text (with) parentheses",
			expectedResult: "text \\(with\\) parentheses",
		},
		{
			name:           "Escape_Tilde",
			text:           "text ~with~ tilde",
			expectedResult: "text \\~with\\~ tilde",
		},
		{
			name:           "Escape_Backtick",
			text:           "text `with` backtick",
			expectedResult: "text \\`with\\` backtick",
		},
		{
			name:           "Escape_Greater_Than",
			text:           "text >with> greater",
			expectedResult: "text \\>with\\> greater",
		},
		{
			name:           "Escape_Hash",
			text:           "text #with# hash",
			expectedResult: "text \\#with\\# hash",
		},
		{
			name:           "Escape_Plus",
			text:           "text +with+ plus",
			expectedResult: "text \\+with\\+ plus",
		},
		{
			name:           "Escape_Minus",
			text:           "text -with- minus",
			expectedResult: "text \\-with\\- minus",
		},
		{
			name:           "Escape_Equals",
			text:           "text =with= equals",
			expectedResult: "text \\=with\\= equals",
		},
		{
			name:           "Escape_Pipe",
			text:           "text |with| pipe",
			expectedResult: "text \\|with\\| pipe",
		},
		{
			name:           "Escape_Braces",
			text:           "text {with} braces",
			expectedResult: "text \\{with\\} braces",
		},
		{
			name:           "Escape_Dot",
			text:           "text .with. dot",
			expectedResult: "text \\.with\\. dot",
		},
		{
			name:           "Escape_Exclamation",
			text:           "text !with! exclamation",
			expectedResult: "text \\!with\\! exclamation",
		},
		{
			name:           "Escape_All_Special_Chars",
			text:           "_*[]()~`>#+-=|{}.!",
			expectedResult: "\\_\\*\\[\\]\\(\\)\\~\\`\\>\\#\\+\\-\\=\\|\\{\\}\\.\\!",
		},
		{
			name:           "Escape_Mixed_Text",
			text:           "Task #123: Fix bug (urgent!)",
			expectedResult: "Task \\#123: Fix bug \\(urgent\\!\\)",
		},
		{
			name:           "Escape_Already_Escaped",
			text:           "text \\_with\\_ escaped",
			expectedResult: "text \\\\_with\\\\_ escaped",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := escapeMarkdownV2(tc.text)

			if result != tc.expectedResult {
				t.Errorf("expected result %q, got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestEscapeMarkdownV2LinkText(t *testing.T) {
	type testCase struct {
		name           string
		text           string
		expectedResult string
	}

	testCases := []testCase{
		{
			name:           "Escape_Link_Text_Empty",
			text:           "",
			expectedResult: "",
		},
		{
			name:           "Escape_Link_Text_Simple",
			text:           "Simple link text",
			expectedResult: "Simple link text",
		},
		{
			name:           "Escape_Link_Text_With_Brackets",
			text:           "text [with] brackets",
			expectedResult: "text [with] brackets",
		},
		{
			name:           "Escape_Link_Text_With_Other_Special_Chars",
			text:           "text _*()~`>#+-=|{}.!",
			expectedResult: "text \\_\\*\\(\\)\\~\\`\\>\\#\\+\\-\\=\\|\\{\\}\\.\\!",
		},
		{
			name:           "Escape_Link_Text_Mixed",
			text:           "Task #123 (urgent!)",
			expectedResult: "Task \\#123 \\(urgent\\!\\)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := escapeMarkdownV2LinkText(tc.text)

			if result != tc.expectedResult {
				t.Errorf("expected result %q, got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestEscapeMarkdownV2URL(t *testing.T) {
	type testCase struct {
		name           string
		url            string
		expectedResult string
	}

	testCases := []testCase{
		{
			name:           "Escape_URL_Empty",
			url:            "",
			expectedResult: "",
		},
		{
			name:           "Escape_URL_Simple",
			url:            "https://example.com",
			expectedResult: "https://example.com",
		},
		{
			name:           "Escape_URL_With_Backslash",
			url:            "https://example.com\\path",
			expectedResult: "https://example.com\\\\path",
		},
		{
			name:           "Escape_URL_With_Parenthesis",
			url:            "https://example.com/path)",
			expectedResult: "https://example.com/path\\)",
		},
		{
			name:           "Escape_URL_With_Both",
			url:            "https://example.com\\path)",
			expectedResult: "https://example.com\\\\path\\)",
		},
		{
			name:           "Escape_URL_With_Multiple_Backslashes",
			url:            "https://example.com\\path\\to\\file",
			expectedResult: "https://example.com\\\\path\\\\to\\\\file",
		},
		{
			name:           "Escape_URL_With_Multiple_Parentheses",
			url:            "https://example.com/path(1)(2)",
			expectedResult: "https://example.com/path(1\\)(2\\)",
		},
		{
			name:           "Escape_URL_Complex",
			url:            "https://youtrack.test/issue/PROJ-123\\test)",
			expectedResult: "https://youtrack.test/issue/PROJ-123\\\\test\\)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := escapeMarkdownV2URL(tc.url)

			if result != tc.expectedResult {
				t.Errorf("expected result %q, got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestGetFieldIcon(t *testing.T) {
	type testCase struct {
		name           string
		field          string
		expectedResult string
	}

	testCases := []testCase{
		{
			name:           "Get_Icon_State",
			field:          State,
			expectedResult: "📊",
		},
		{
			name:           "Get_Icon_Priority",
			field:          Priority,
			expectedResult: "⚡",
		},
		{
			name:           "Get_Icon_Assignee",
			field:          Assignee,
			expectedResult: "👤",
		},
		{
			name:           "Get_Icon_Comment",
			field:          Comment,
			expectedResult: "💬",
		},
		{
			name:           "Get_Icon_Unknown_Field",
			field:          "UnknownField",
			expectedResult: "📝",
		},
		{
			name:           "Get_Icon_Empty_Field",
			field:          "",
			expectedResult: "📝",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getFieldIcon(tc.field)

			if result != tc.expectedResult {
				t.Errorf("expected result %q, got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestExtractChangesTelegram(t *testing.T) {
	type testCase struct {
		name           string
		changes        []parser.YoutrackChange
		expectedResult string
		checkContains  []string
	}

	testCases := []testCase{
		{
			name:           "Extract_Changes_Empty",
			changes:        []parser.YoutrackChange{},
			expectedResult: "",
		},
		{
			name:           "Extract_Changes_Nil",
			changes:        nil,
			expectedResult: "",
		},
		{
			name: "Extract_Changes_State",
			changes: []parser.YoutrackChange{
				{
					Field:    State,
					OldValue: []byte(`{"name": "To Do", "presentation": "К выполнению"}`),
					NewValue: []byte(`{"name": "In Progress", "presentation": "В работе"}`),
				},
			},
			checkContains: []string{"📊", "*Состояние:*", "К выполнению", "→", "В работе"},
		},
		{
			name: "Extract_Changes_Priority",
			changes: []parser.YoutrackChange{
				{
					Field:    Priority,
					OldValue: []byte(`{"name": "Low", "presentation": "Низкий"}`),
					NewValue: []byte(`{"name": "High", "presentation": "Высокий"}`),
				},
			},
			checkContains: []string{"⚡", "*Приоритет:*", "Низкий", "→", "Высокий"},
		},
		{
			name: "Extract_Changes_Assignee",
			changes: []parser.YoutrackChange{
				{
					Field:    Assignee,
					OldValue: []byte(`{"fullName": "John Doe", "login": "john"}`),
					NewValue: []byte(`{"fullName": "Jane Smith", "login": "jane"}`),
				},
			},
			checkContains: []string{"👤", "*Назначена:*", "John Doe", "→", "Jane Smith"},
		},
		{
			name: "Extract_Changes_Comment",
			changes: []parser.YoutrackChange{
				{
					Field:    Comment,
					OldValue: []byte(`null`),
					NewValue: []byte(`{"text": "Test comment", "mentionedUsers": []}`),
				},
			},
			checkContains: []string{"💬", "*Комментарий:*", "Test comment"},
		},
		{
			name: "Extract_Changes_Multiple",
			changes: []parser.YoutrackChange{
				{
					Field:    State,
					OldValue: []byte(`"To Do"`),
					NewValue: []byte(`"In Progress"`),
				},
				{
					Field:    Priority,
					OldValue: []byte(`"Low"`),
					NewValue: []byte(`"High"`),
				},
			},
			checkContains: []string{"📊", "*Состояние:*", "⚡", "*Приоритет:*"},
		},
		{
			name: "Extract_Changes_Comment_With_Special_Chars",
			changes: []parser.YoutrackChange{
				{
					Field:    Comment,
					OldValue: []byte(`null`),
					NewValue: []byte(`{"text": "Comment with *asterisk* and _underscore_", "mentionedUsers": []}`),
				},
			},
			checkContains: []string{"💬", "*Комментарий:*", "Comment with \\*asterisk\\* and \\_underscore\\_"},
		},
		{
			name: "Extract_Changes_Unknown_Field",
			changes: []parser.YoutrackChange{
				{
					Field:    "UnknownField",
					OldValue: []byte(`"old"`),
					NewValue: []byte(`"new"`),
				},
			},
			checkContains: []string{"📝", "*UnknownField:*"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractChangesTelegram(tc.changes)

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
		})
	}
}

func TestFormatTelegram(t *testing.T) {
	type testCase struct {
		name             string
		payload          *parser.YoutrackWebhookPayload
		expectedResult   string
		checkContains    []string
		checkNotContains []string
	}

	projectName := "TestProject"
	projectPresentation := "Test Project"
	issueSummary := "Test Issue Summary"
	issueURL := "https://youtrack.test/issue/PROJ-123"
	stateName := "In Progress"
	statePresentation := "В работе"
	priorityName := "High"
	priorityPresentation := "Высокий"
	assigneeFullName := "John Doe"
	assigneeLogin := "john"
	updaterFullName := "Jane Smith"
	updaterLogin := "jane"

	testCases := []testCase{
		{
			name: "Format_Telegram_Basic",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{
					Name:         &projectName,
					Presentation: &projectPresentation,
				},
				Issue: parser.YoutrackIssue{
					Summary: issueSummary,
					URL:     issueURL,
					State: &parser.YoutrackFieldValue{
						Name:         &stateName,
						Presentation: &statePresentation,
					},
					Priority: &parser.YoutrackFieldValue{
						Name:         &priorityName,
						Presentation: &priorityPresentation,
					},
					Assignee: &parser.YoutrackUser{
						FullName: &assigneeFullName,
						Login:    &assigneeLogin,
					},
				},
				Updater: &parser.YoutrackUser{
					FullName: &updaterFullName,
					Login:    &updaterLogin,
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    State,
						OldValue: []byte(`"To Do"`),
						NewValue: []byte(`"In Progress"`),
					},
				},
			},
			checkContains: []string{
				"*📊 Изменен статус задачи*",
				"*📁 Проект:*",
				"TestProject",
				"*📋 Задача:*",
				"Test Issue Summary",
				"🔗 *Ссылка:*",
				"*📊 Состояние:*",
				"В работе",
				"*⚡️ Приоритет:*",
				"Высокий",
				"*👤 Назначена:*",
				"John Doe",
				"*✏️ Автор изменения:*",
				"Jane Smith",
				"🔄 *Изменения:*",
				"📊",
				"*Состояние:*",
			},
		},
		{
			name: "Format_Telegram_With_Assignee_Change",
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
					FullName: &updaterFullName,
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    Assignee,
						OldValue: []byte(`{"fullName": "Old User"}`),
						NewValue: []byte(`{"fullName": "New User"}`),
					},
				},
			},
			checkContains: []string{
				"*👤 Изменен исполнитель задачи*",
				"*👤 Назначена:*",
				"Old User",
				"→",
				"New User",
			},
		},
		{
			name: "Format_Telegram_With_Comment_Change",
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
					FullName: &updaterFullName,
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    Comment,
						OldValue: []byte(`null`),
						NewValue: []byte(`{"text": "New comment text", "mentionedUsers": []}`),
					},
				},
			},
			checkContains: []string{
				"*💬 Добавлен комментарий*",
				"💬",
				"*Комментарий:*",
				"New comment text",
			},
		},
		{
			name: "Format_Telegram_With_Priority_Change",
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
					FullName: &updaterFullName,
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    Priority,
						OldValue: []byte(`"Low"`),
						NewValue: []byte(`"High"`),
					},
				},
			},
			checkContains: []string{
				"*⚡️ Изменен приоритет задачи*",
				"⚡",
				"*Приоритет:*",
			},
		},
		{
			name: "Format_Telegram_With_Special_Chars",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{
					Name: stringPtr("Project_With*Special[Chars]"),
				},
				Issue: parser.YoutrackIssue{
					Summary: "Issue with *asterisk* and _underscore_",
					URL:     "https://youtrack.test/issue/PROJ-123\\test)",
					State: &parser.YoutrackFieldValue{
						Name: stringPtr("State (with) parentheses"),
					},
					Priority: &parser.YoutrackFieldValue{
						Name: stringPtr("Priority #high"),
					},
					Assignee: &parser.YoutrackUser{
						FullName: stringPtr("User >with< special"),
					},
				},
				Updater: &parser.YoutrackUser{
					FullName: stringPtr("Updater ~with~ tilde"),
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    State,
						OldValue: []byte(`"Old"`),
						NewValue: []byte(`"New"`),
					},
				},
			},
			checkContains: []string{
				"Project\\_With\\*Special\\[Chars\\]",
				"Issue with \\*asterisk\\* and \\_underscore\\_",
				"State \\(with\\) parentheses",
				"Priority \\#high",
				"User \\>with< special",
				"Updater \\~with\\~ tilde",
			},
		},
		{
			name: "Format_Telegram_With_Multiple_Changes",
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
					FullName: &updaterFullName,
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    State,
						OldValue: []byte(`"To Do"`),
						NewValue: []byte(`"In Progress"`),
					},
					{
						Field:    Priority,
						OldValue: []byte(`"Low"`),
						NewValue: []byte(`"High"`),
					},
					{
						Field:    Assignee,
						OldValue: []byte(`{"fullName": "Old"}`),
						NewValue: []byte(`{"fullName": "New"}`),
					},
				},
			},
			checkContains: []string{
				"*📊 Изменен статус задачи*",
				"*⚡️ Изменен приоритет задачи*",
				"*👤 Изменен исполнитель задачи*",
			},
		},
		{
			name: "Format_Telegram_With_No_Changes",
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
					FullName: &updaterFullName,
				},
				Changes: []parser.YoutrackChange{},
			},
			checkContains: []string{
				"*📁 Проект:*",
				"*📋 Задача:*",
			},
			checkNotContains: []string{
				"🔄 *Изменения:*",
			},
		},
		{
			name: "Format_Telegram_With_Nil_Fields",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{
					Name: &projectName,
				},
				Issue: parser.YoutrackIssue{
					Summary:  issueSummary,
					URL:      issueURL,
					State:    nil,
					Priority: nil,
					Assignee: nil,
				},
				Updater: nil,
				Changes: []parser.YoutrackChange{
					{
						Field:    State,
						OldValue: []byte(`"Old"`),
						NewValue: []byte(`"New"`),
					},
				},
			},
			checkContains: []string{
				"*📊 Состояние:*",
				"*⚡️ Приоритет:*",
				"*👤 Назначена:*",
				"*✏️ Автор изменения:*",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatTelegram(tc.payload)

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
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
