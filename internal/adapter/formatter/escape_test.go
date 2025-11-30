package formatter

import "testing"

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
