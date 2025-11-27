package formatter

import (
	"encoding/json"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"strings"
	"testing"
)

func TestExtractChangesVKTeams(t *testing.T) {
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
			result := extractChangesVKTeams(tc.changes)

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

func TestExtractChangeValueVKTeams(t *testing.T) {
	type testCase struct {
		name           string
		value          json.RawMessage
		field          string
		expectedResult string
		checkContains  []string
	}

	testCases := []testCase{
		{
			name:           "Extract_Change_Value_Empty",
			value:          []byte(``),
			field:          "Unknown",
			expectedResult: nullValueString,
		},
		{
			name:           "Extract_Change_Value_Null",
			value:          []byte(`null`),
			field:          "Unknown",
			expectedResult: nullValueString,
		},
		{
			name:           "Extract_Change_Value_Empty_String",
			value:          []byte(`""`),
			field:          "Unknown",
			expectedResult: "",
		},
		{
			name:          "Extract_Change_Value_State",
			value:         []byte(`{"name": "In Progress", "presentation": "В работе"}`),
			field:         State,
			checkContains: []string{"В работе"},
		},
		{
			name:          "Extract_Change_Value_Priority",
			value:         []byte(`{"name": "High", "presentation": "Высокий"}`),
			field:         Priority,
			checkContains: []string{"Высокий"},
		},
		{
			name:          "Extract_Change_Value_Assignee",
			value:         []byte(`{"fullName": "John Doe", "login": "john"}`),
			field:         Assignee,
			checkContains: []string{"John Doe"},
		},
		{
			name:          "Extract_Change_Value_Comment",
			value:         []byte(`{"text": "Test comment", "mentionedUsers": []}`),
			field:         Comment,
			checkContains: []string{"Test comment"},
		},
		{
			name:           "Extract_Change_Value_String",
			value:          []byte(`"Simple string"`),
			field:          "Unknown",
			expectedResult: "Simple string",
		},
		{
			name:          "Extract_Change_Value_Object_With_Name",
			value:         []byte(`{"name": "TestName", "value": "TestValue"}`),
			field:         "Unknown",
			checkContains: []string{"TestName"},
		},
		{
			name:          "Extract_Change_Value_Object_With_Value",
			value:         []byte(`{"value": "TestValue"}`),
			field:         "Unknown",
			checkContains: []string{"TestValue"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractChangeValueVKTeams(tc.value, tc.field)

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

func TestExtractCommentTextVKTeams(t *testing.T) {
	type testCase struct {
		name             string
		comment          parser.YoutrackCommentValue
		expectedResult   string
		checkContains    []string
		checkNotContains []string
	}

	email1 := "user1@example.com"
	email2 := "user2@example.com"
	login1 := "user1"
	login2 := "user2"
	fullName1 := "User One"
	fullName2 := "User Two"

	testCases := []testCase{
		{
			name: "Extract_Comment_Text_Empty",
			comment: parser.YoutrackCommentValue{
				Text:           "",
				MentionedUsers: []parser.YoutrackUser{},
			},
			expectedResult: "",
		},
		{
			name: "Extract_Comment_Text_Simple",
			comment: parser.YoutrackCommentValue{
				Text:           "Simple comment text",
				MentionedUsers: []parser.YoutrackUser{},
			},
			expectedResult: "Simple comment text",
		},
		{
			name: "Extract_Comment_Text_With_Special_Chars",
			comment: parser.YoutrackCommentValue{
				Text:           "Comment with *asterisk* and _underscore_",
				MentionedUsers: []parser.YoutrackUser{},
			},
			expectedResult: "Comment with *asterisk* and _underscore_",
		},
		{
			name: "Extract_Comment_Text_With_Email_Mention",
			comment: parser.YoutrackCommentValue{
				Text: "Comment with mention",
				MentionedUsers: []parser.YoutrackUser{
					{
						Email:    &email1,
						FullName: &fullName1,
						Login:    &login1,
					},
				},
			},
			checkContains:    []string{"Comment with mention", "\n[Упомянуты:", email1},
			checkNotContains: []string{"@" + login1},
		},
		{
			name: "Extract_Comment_Text_With_Login_Mention_No_Email",
			comment: parser.YoutrackCommentValue{
				Text: "Comment with login mention",
				MentionedUsers: []parser.YoutrackUser{
					{
						FullName: &fullName1,
						Login:    &login1,
					},
				},
			},
			checkContains: []string{"Comment with login mention", "\n[Упомянуты:", "@" + login1},
		},
		{
			name: "Extract_Comment_Text_With_Multiple_Mentions_Email_Priority",
			comment: parser.YoutrackCommentValue{
				Text: "Comment with multiple mentions",
				MentionedUsers: []parser.YoutrackUser{
					{
						Email:    &email1,
						FullName: &fullName1,
						Login:    &login1,
					},
					{
						Email:    &email2,
						FullName: &fullName2,
						Login:    &login2,
					},
				},
			},
			checkContains:    []string{"Comment with multiple mentions", "\n[Упомянуты:", email1, email2},
			checkNotContains: []string{"@" + login1, "@" + login2},
		},
		{
			name: "Extract_Comment_Text_With_Mixed_Email_And_Login",
			comment: parser.YoutrackCommentValue{
				Text: "Comment with mixed mentions",
				MentionedUsers: []parser.YoutrackUser{
					{
						Email:    &email1,
						FullName: &fullName1,
						Login:    &login1,
					},
					{
						FullName: &fullName2,
						Login:    &login2,
					},
				},
			},
			checkContains: []string{"Comment with mixed mentions", "\n[Упомянуты:", email1, "@" + login2},
		},
		{
			name: "Extract_Comment_Text_With_Login_Only_No_FullName",
			comment: parser.YoutrackCommentValue{
				Text: "Comment with login only",
				MentionedUsers: []parser.YoutrackUser{
					{
						Login: &login1,
					},
				},
			},
			checkContains: []string{"Comment with login only", "\n[Упомянуты:", "@" + login1},
		},
		{
			name: "Extract_Comment_Text_With_Empty_Email_And_Login",
			comment: parser.YoutrackCommentValue{
				Text: "Comment text",
				MentionedUsers: []parser.YoutrackUser{
					{
						FullName: &fullName1,
					},
				},
			},
			expectedResult:   "Comment text",
			checkNotContains: []string{"[Упомянуты:"},
		},
		{
			name: "Extract_Comment_Text_With_Empty_MentionedUsers",
			comment: parser.YoutrackCommentValue{
				Text:           "Comment text",
				MentionedUsers: []parser.YoutrackUser{},
			},
			expectedResult:   "Comment text",
			checkNotContains: []string{"[Упомянуты:"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractCommentTextVKTeams(tc.comment)

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

func TestFormatVKTeams(t *testing.T) {
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
			name: "Format_VKTeams_Basic",
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
			name: "Format_VKTeams_With_Assignee_Change",
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
			name: "Format_VKTeams_With_Comment_Change",
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
			name: "Format_VKTeams_With_Comment_With_Email_Mention",
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
						NewValue: []byte(`{"text": "Comment with mention", "mentionedUsers": [{"email": "user1@example.com", "fullName": "User One", "login": "user1"}]}`),
					},
				},
			},
			checkContains: []string{
				"*💬 Добавлен комментарий*",
				"Comment with mention",
				"\\[Упомянуты:",
				"user1@example\\.com",
			},
			checkNotContains: []string{"@user1"},
		},
		{
			name: "Format_VKTeams_With_Comment_With_Login_Mention",
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
						NewValue: []byte(`{"text": "Comment with login mention", "mentionedUsers": [{"fullName": "User One", "login": "user1"}]}`),
					},
				},
			},
			checkContains: []string{
				"*💬 Добавлен комментарий*",
				"Comment with login mention",
				"\\[Упомянуты:",
				"@user1",
			},
		},
		{
			name: "Format_VKTeams_With_Priority_Change",
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
			name: "Format_VKTeams_With_Special_Chars",
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
			name: "Format_VKTeams_With_Multiple_Changes",
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
			name: "Format_VKTeams_With_No_Changes",
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
			name: "Format_VKTeams_With_Nil_Fields",
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
		{
			name: "Format_VKTeams_With_Comment_Multiple_Mentions",
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
						NewValue: []byte(`{"text": "Comment with multiple mentions", "mentionedUsers": [{"email": "user1@example.com", "fullName": "User One", "login": "user1"}, {"fullName": "User Two", "login": "user2"}]}`),
					},
				},
			},
			checkContains: []string{
				"Comment with multiple mentions",
				"\\[Упомянуты:",
				"user1@example\\.com",
				"@user2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatVKTeams(tc.payload)

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
