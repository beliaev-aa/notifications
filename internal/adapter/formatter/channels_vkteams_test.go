package formatter

import (
	"encoding/json"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"strings"
	"testing"
)

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
			name:          "Extract_Change_Value_Comment_Simple",
			value:         []byte(`{"text": "Test comment", "mentionedUsers": []}`),
			field:         Comment,
			checkContains: []string{"Test comment"},
		},
		{
			name:           "Extract_Change_Value_Comment_Invalid_JSON",
			value:          []byte(`{"invalid": json}`),
			field:          Comment,
			expectedResult: nullValueString,
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
			name:           "Extract_Change_Value_Object_No_Name_No_Value",
			value:          []byte(`{"other": "field"}`),
			field:          "Unknown",
			expectedResult: nullValueString,
		},
		{
			name:           "Extract_Change_Value_Invalid_JSON",
			value:          []byte(`{invalid json}`),
			field:          "Unknown",
			expectedResult: nullValueString,
		},
		{
			name:           "Extract_Change_Value_State_Invalid_JSON",
			value:          []byte(`{invalid}`),
			field:          State,
			expectedResult: nullValueString,
		},
		{
			name:           "Extract_Change_Value_Priority_Invalid_JSON",
			value:          []byte(`{invalid}`),
			field:          Priority,
			expectedResult: nullValueString,
		},
		{
			name:           "Extract_Change_Value_Assignee_Invalid_JSON",
			value:          []byte(`{invalid}`),
			field:          Assignee,
			expectedResult: nullValueString,
		},
		{
			name:           "Extract_Change_Value_Comment_Invalid_JSON",
			value:          []byte(`{invalid}`),
			field:          Comment,
			expectedResult: nullValueString,
		},
		{
			name:          "Extract_Change_Value_Unknown_Field_Object_With_Value_No_Name",
			value:         []byte(`{"value": "TestValue", "other": "field"}`),
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
	login1 := "user1"
	login2 := "user2"
	fullName1 := "User One"

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
			checkNotContains: []string{"@" + login1, fullName1},
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
						Login: &login2,
					},
				},
			},
			checkContains: []string{"Comment with mixed mentions", "\n[Упомянуты:", email1, "@" + login2},
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
						OldValue: []byte(`{"name": "To Do", "presentation": "К выполнению"}`),
						NewValue: []byte(`{"name": "In Progress", "presentation": "В работе"}`),
					},
				},
			},
			checkContains: []string{
				"*📊 Изменен статус задачи*",
				"*📁 Проект:*",
				"TestProject",
				"*👤 Назначена:*",
				"@john",
				"*✏️ Автор изменения:*",
				"Jane Smith",
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
				"@\\[user1@example\\.com\\]",
			},
			checkNotContains: []string{
				"User One",
				"@user1",
			},
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
			name: "Format_VKTeams_With_Comment_Mixed_Email_And_Login",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{
					Name: &projectName,
				},
				Issue: parser.YoutrackIssue{
					Summary: issueSummary,
					URL:     issueURL,
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    Comment,
						OldValue: []byte(`null`),
						NewValue: []byte(`{"text": "Comment", "mentionedUsers": [{"email": "user1@example.com", "login": "user1"}, {"login": "user2"}]}`),
					},
				},
			},
			checkContains: []string{
				"@\\[user1@example\\.com\\]",
				"@user2",
			},
		},
		{
			name: "Format_VKTeams_With_Assignee_Email",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{
					Name: &projectName,
				},
				Issue: parser.YoutrackIssue{
					Summary: issueSummary,
					URL:     issueURL,
					Assignee: &parser.YoutrackUser{
						FullName: &assigneeFullName,
						Email:    stringPtr("john@example.com"),
						Login:    &assigneeLogin,
					},
				},
				Changes: []parser.YoutrackChange{},
			},
			checkContains: []string{
				"@\\[john@example\\.com\\]",
			},
			checkNotContains: []string{
				"@john",
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
					Priority: &parser.YoutrackFieldValue{
						Name: &priorityName,
					},
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    Priority,
						OldValue: []byte(`{"name": "Low", "presentation": "Низкий"}`),
						NewValue: []byte(`{"name": "High", "presentation": "Высокий"}`),
					},
				},
			},
			checkContains: []string{
				"Изменен приоритет задачи",
				"*⚡️ Приоритет:*",
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
					Assignee: &parser.YoutrackUser{
						Login: &assigneeLogin,
					},
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    Assignee,
						OldValue: []byte(`{"login": "old_user"}`),
						NewValue: []byte(`{"login": "new_user"}`),
					},
				},
			},
			checkContains: []string{
				"*👤 Изменен исполнитель задачи*",
				"*👤 Назначена:*",
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
				},
				Changes: []parser.YoutrackChange{},
			},
			checkContains: []string{
				"*📁 Проект:*",
				"*📋 Задача:*",
			},
			checkNotContains: []string{
				"*💬 Добавлен комментарий*",
				"*📊 Изменен статус задачи*",
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
						OldValue: []byte(`{"name": "Old", "presentation": "Старое"}`),
						NewValue: []byte(`{"name": "New", "presentation": "Новое"}`),
					},
				},
			},
			checkContains: []string{
				"*📊 Состояние:*",
				"*⚡️ Приоритет:*",
				"*👤 Назначена:*",
			},
		},
		{
			name: "Format_VKTeams_With_State_Change_Shows_Changed_Value",
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
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    State,
						OldValue: []byte(`{"name": "Old", "presentation": "Старое"}`),
						NewValue: []byte(`{"name": "New", "presentation": "Новое"}`),
					},
				},
			},
			checkContains: []string{
				"*📊 Состояние:*",
				"Старое → Новое",
			},
		},
		{
			name: "Format_VKTeams_With_Priority_Change_Shows_Changed_Value",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{
					Name: &projectName,
				},
				Issue: parser.YoutrackIssue{
					Summary: issueSummary,
					URL:     issueURL,
					Priority: &parser.YoutrackFieldValue{
						Name: &priorityName,
					},
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    Priority,
						OldValue: []byte(`{"name": "Low", "presentation": "Низкий"}`),
						NewValue: []byte(`{"name": "High", "presentation": "Высокий"}`),
					},
				},
			},
			checkContains: []string{
				"*⚡️ Приоритет:*",
				"Низкий → Высокий",
			},
		},
		{
			name: "Format_VKTeams_With_Assignee_Change_Shows_Changed_Value",
			payload: &parser.YoutrackWebhookPayload{
				Project: &parser.YoutrackFieldValue{
					Name: &projectName,
				},
				Issue: parser.YoutrackIssue{
					Summary: issueSummary,
					URL:     issueURL,
					Assignee: &parser.YoutrackUser{
						Login: &assigneeLogin,
					},
				},
				Changes: []parser.YoutrackChange{
					{
						Field:    Assignee,
						OldValue: []byte(`{"login": "old_user"}`),
						NewValue: []byte(`{"login": "new_user"}`),
					},
				},
			},
			checkContains: []string{
				"*👤 Назначена:*",
				"old\\_user →",
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

func TestVKTeamsMentionFormatter_FormatMention(t *testing.T) {
	type testCase struct {
		name           string
		user           parser.YoutrackUser
		expectedResult string
		checkContains  []string
	}

	email := "john@example.com"
	login := "john"
	emptyString := ""

	testCases := []testCase{
		{
			name: "Format_Mention_With_Email",
			user: parser.YoutrackUser{
				Email: &email,
				Login: &login,
			},
			checkContains: []string{"@[", email, "]"},
		},
		{
			name: "Format_Mention_With_Login_No_Email",
			user: parser.YoutrackUser{
				Login: &login,
			},
			checkContains: []string{"@", login},
		},
		{
			name: "Format_Mention_All_Empty",
			user: parser.YoutrackUser{
				Email: &emptyString,
				Login: &emptyString,
			},
			expectedResult: "",
		},
		{
			name:           "Format_Mention_All_Nil",
			user:           parser.YoutrackUser{},
			expectedResult: "",
		},
	}

	formatter := &VKTeamsMentionFormatter{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatter.FormatMention(tc.user)

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
