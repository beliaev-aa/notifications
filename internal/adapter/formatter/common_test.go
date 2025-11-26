package formatter

import (
	"encoding/json"
	"github.com/beliaev-aa/notifications/internal/domain/port/parser"
	"strings"
	"testing"
)

func TestTranslateFieldName(t *testing.T) {
	type testCase struct {
		name           string
		field          string
		expectedResult string
	}

	testCases := []testCase{
		{
			name:           "Translate_Assignee",
			field:          Assignee,
			expectedResult: "Назначена",
		},
		{
			name:           "Translate_Comment",
			field:          Comment,
			expectedResult: "Комментарий",
		},
		{
			name:           "Translate_Priority",
			field:          Priority,
			expectedResult: "Приоритет",
		},
		{
			name:           "Translate_State",
			field:          State,
			expectedResult: "Состояние",
		},
		{
			name:           "Translate_Unknown_Field",
			field:          "UnknownField",
			expectedResult: "UnknownField",
		},
		{
			name:           "Translate_Empty_Field",
			field:          "",
			expectedResult: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := translateFieldName(tc.field)

			if result != tc.expectedResult {
				t.Errorf("expected result %q, got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestExtractFieldValue(t *testing.T) {
	type testCase struct {
		name           string
		field          *parser.YoutrackFieldValue
		expectedResult string
	}

	fieldName := "TestName"
	fieldPresentation := "Test Presentation"
	emptyString := ""

	testCases := []testCase{
		{
			name: "Extract_Field_Value_With_Presentation",
			field: &parser.YoutrackFieldValue{
				Name:         &fieldName,
				Presentation: &fieldPresentation,
			},
			expectedResult: fieldPresentation,
		},
		{
			name: "Extract_Field_Value_With_Name_Only",
			field: &parser.YoutrackFieldValue{
				Name:         &fieldName,
				Presentation: nil,
			},
			expectedResult: fieldName,
		},
		{
			name: "Extract_Field_Value_With_Empty_Presentation_Uses_Name",
			field: &parser.YoutrackFieldValue{
				Name:         &fieldName,
				Presentation: &emptyString,
			},
			expectedResult: fieldName,
		},
		{
			name: "Extract_Field_Value_With_Empty_Name_Uses_Presentation",
			field: &parser.YoutrackFieldValue{
				Name:         &emptyString,
				Presentation: &fieldPresentation,
			},
			expectedResult: fieldPresentation,
		},
		{
			name:           "Extract_Field_Value_Nil",
			field:          nil,
			expectedResult: "",
		},
		{
			name: "Extract_Field_Value_All_Nil",
			field: &parser.YoutrackFieldValue{
				Name:         nil,
				Presentation: nil,
			},
			expectedResult: "",
		},
		{
			name: "Extract_Field_Value_Both_Empty",
			field: &parser.YoutrackFieldValue{
				Name:         &emptyString,
				Presentation: &emptyString,
			},
			expectedResult: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractFieldValue(tc.field)

			if result != tc.expectedResult {
				t.Errorf("expected result %q, got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestExtractUserName(t *testing.T) {
	type testCase struct {
		name           string
		user           *parser.YoutrackUser
		expectedResult string
	}

	fullName := "John Doe"
	login := "login"
	email := "john@example.com"
	emptyString := ""

	testCases := []testCase{
		{
			name: "Extract_User_Name_With_FullName",
			user: &parser.YoutrackUser{
				FullName: &fullName,
				Login:    &login,
				Email:    &email,
			},
			expectedResult: fullName,
		},
		{
			name: "Extract_User_Name_With_Login_Only",
			user: &parser.YoutrackUser{
				FullName: nil,
				Login:    &login,
			},
			expectedResult: login,
		},
		{
			name: "Extract_User_Name_With_Empty_FullName_Uses_Login",
			user: &parser.YoutrackUser{
				FullName: &emptyString,
				Login:    &login,
			},
			expectedResult: login,
		},
		{
			name: "Extract_User_Name_With_Empty_Login_Uses_FullName",
			user: &parser.YoutrackUser{
				FullName: &fullName,
				Login:    &emptyString,
			},
			expectedResult: fullName,
		},
		{
			name:           "Extract_User_Name_Nil",
			user:           nil,
			expectedResult: "",
		},
		{
			name: "Extract_User_Name_All_Nil",
			user: &parser.YoutrackUser{
				FullName: nil,
				Login:    nil,
				Email:    nil,
			},
			expectedResult: "",
		},
		{
			name: "Extract_User_Name_Both_Empty",
			user: &parser.YoutrackUser{
				FullName: &emptyString,
				Login:    &emptyString,
			},
			expectedResult: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractUserName(tc.user)

			if result != tc.expectedResult {
				t.Errorf("expected result %q, got: %q", tc.expectedResult, result)
			}
		})
	}
}

func TestExtractChangeValue(t *testing.T) {
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
			name:           "Extract_Change_Value_State_With_FieldValue",
			value:          []byte(`{"name": "In Progress", "presentation": "В работе"}`),
			field:          State,
			expectedResult: "В работе",
		},
		{
			name:           "Extract_Change_Value_State_With_Name_Only",
			value:          []byte(`{"name": "In Progress"}`),
			field:          State,
			expectedResult: "In Progress",
		},
		{
			name:           "Extract_Change_Value_Priority_With_FieldValue",
			value:          []byte(`{"name": "High", "presentation": "Высокий"}`),
			field:          Priority,
			expectedResult: "Высокий",
		},
		{
			name:           "Extract_Change_Value_Assignee_With_User",
			value:          []byte(`{"fullName": "John Doe", "login": "login"}`),
			field:          Assignee,
			expectedResult: "John Doe",
		},
		{
			name:           "Extract_Change_Value_Assignee_With_Login_Only",
			value:          []byte(`{"login": "login"}`),
			field:          Assignee,
			expectedResult: "login",
		},
		{
			name:           "Extract_Change_Value_Comment_With_Text",
			value:          []byte(`{"text": "Test comment", "mentionedUsers": []}`),
			field:          Comment,
			expectedResult: "Test comment",
		},
		{
			name:          "Extract_Change_Value_Comment_With_Mentions",
			value:         []byte(`{"text": "Test comment", "mentionedUsers": [{"fullName": "John Doe"}]}`),
			field:         Comment,
			checkContains: []string{"Test comment", "Упомянуты:", "John Doe"},
		},
		{
			name:           "Extract_Change_Value_Unknown_Field_String",
			value:          []byte(`"simple string"`),
			field:          "Unknown",
			expectedResult: "simple string",
		},
		{
			name:           "Extract_Change_Value_Unknown_Field_Object_With_Name",
			value:          []byte(`{"name": "Test Name"}`),
			field:          "Unknown",
			expectedResult: "Test Name",
		},
		{
			name:           "Extract_Change_Value_Unknown_Field_Object_With_Value",
			value:          []byte(`{"value": "Test Value"}`),
			field:          "Unknown",
			expectedResult: "Test Value",
		},
		{
			name:           "Extract_Change_Value_Invalid_JSON",
			value:          []byte(`{invalid json}`),
			field:          State,
			expectedResult: nullValueString,
		},
		{
			name:           "Extract_Change_Value_State_Invalid_JSON",
			value:          []byte(`{invalid}`),
			field:          State,
			expectedResult: nullValueString,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractChangeValue(tc.value, tc.field)

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

func TestExtractCommentText(t *testing.T) {
	type testCase struct {
		name             string
		comment          parser.YoutrackCommentValue
		expectedResult   string
		checkContains    []string
		checkNotContains []string
	}

	fullName1 := "John Doe"
	fullName2 := "Jane Smith"
	login1 := "login"
	login2 := "other_login"

	testCases := []testCase{
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
				Text:           "Comment with \\* asterisk and \\~ tilde",
				MentionedUsers: []parser.YoutrackUser{},
			},
			expectedResult:   "Comment with * asterisk and ~ tilde",
			checkNotContains: []string{"\\*", "\\~"},
		},
		{
			name: "Extract_Comment_Text_With_All_Special_Chars",
			comment: parser.YoutrackCommentValue{
				Text:           "Test \\* \\~ \\` \\> \\| chars",
				MentionedUsers: []parser.YoutrackUser{},
			},
			expectedResult:   "Test * ~ ` > | chars",
			checkNotContains: []string{"\\*", "\\~", "\\`", "\\>", "\\|"},
		},
		{
			name: "Extract_Comment_Text_With_Single_Mention",
			comment: parser.YoutrackCommentValue{
				Text: "Comment with mention",
				MentionedUsers: []parser.YoutrackUser{
					{
						FullName: &fullName1,
						Login:    &login1,
					},
				},
			},
			checkContains: []string{"Comment with mention", "Упомянуты:", "John Doe"},
		},
		{
			name: "Extract_Comment_Text_With_Multiple_Mentions",
			comment: parser.YoutrackCommentValue{
				Text: "Comment with multiple mentions",
				MentionedUsers: []parser.YoutrackUser{
					{
						FullName: &fullName1,
						Login:    &login1,
					},
					{
						FullName: &fullName2,
						Login:    &login2,
					},
				},
			},
			checkContains: []string{"Comment with multiple mentions", "Упомянуты:", "John Doe", "Jane Smith"},
		},
		{
			name: "Extract_Comment_Text_With_Mention_Login_Only",
			comment: parser.YoutrackCommentValue{
				Text: "Comment with login mention",
				MentionedUsers: []parser.YoutrackUser{
					{
						FullName: nil,
						Login:    &login1,
					},
				},
			},
			checkContains: []string{"Comment with login mention", "Упомянуты:", "login"},
		},
		{
			name: "Extract_Comment_Text_With_Empty_Mention",
			comment: parser.YoutrackCommentValue{
				Text: "Comment with empty mention",
				MentionedUsers: []parser.YoutrackUser{
					{
						FullName: nil,
						Login:    nil,
					},
				},
			},
			expectedResult:   "Comment with empty mention",
			checkNotContains: []string{"Упомянуты:"},
		},
		{
			name: "Extract_Comment_Text_Empty_Text",
			comment: parser.YoutrackCommentValue{
				Text:           "",
				MentionedUsers: []parser.YoutrackUser{},
			},
			expectedResult: "",
		},
		{
			name: "Extract_Comment_Text_With_Mixed_Mentions",
			comment: parser.YoutrackCommentValue{
				Text: "Comment text",
				MentionedUsers: []parser.YoutrackUser{
					{
						FullName: &fullName1,
					},
					{
						FullName: nil,
						Login:    nil,
					},
					{
						Login: &login2,
					},
				},
			},
			checkContains:    []string{"Comment text", "Упомянуты:", "John Doe", "other_login"},
			checkNotContains: []string{", ,"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractCommentText(tc.comment)

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

func TestExtractChangeValue_EdgeCases(t *testing.T) {
	type testCase struct {
		name           string
		value          json.RawMessage
		field          string
		expectedResult string
		checkContains  []string
	}

	testCases := []testCase{
		{
			name:           "Extract_Change_Value_Whitespace_Only",
			value:          []byte(`   `),
			field:          "Unknown",
			expectedResult: nullValueString,
		},
		{
			name:           "Extract_Change_Value_State_With_Empty_Object",
			value:          []byte(`{}`),
			field:          State,
			expectedResult: "",
		},
		{
			name:           "Extract_Change_Value_Assignee_With_Empty_Object",
			value:          []byte(`{}`),
			field:          Assignee,
			expectedResult: "",
		},
		{
			name:           "Extract_Change_Value_Comment_With_Empty_Text",
			value:          []byte(`{"text": "", "mentionedUsers": []}`),
			field:          Comment,
			expectedResult: "",
		},
		{
			name:           "Extract_Change_Value_Comment_With_Null_Text",
			value:          []byte(`{"text": null, "mentionedUsers": []}`),
			field:          Comment,
			expectedResult: "",
		},
		{
			name:           "Extract_Change_Value_Unknown_Field_Object_Empty",
			value:          []byte(`{}`),
			field:          "Unknown",
			expectedResult: nullValueString,
		},
		{
			name:           "Extract_Change_Value_Unknown_Field_Object_No_Name_No_Value",
			value:          []byte(`{"other": "field"}`),
			field:          "Unknown",
			expectedResult: nullValueString,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractChangeValue(tc.value, tc.field)

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
