package formatter

import "strings"

// getFieldIcon возвращает иконку для поля
func getFieldIcon(field string) string {
	icons := map[string]string{
		State:    "📊",
		Priority: "⚡",
		Assignee: "👤",
		Comment:  "💬",
	}
	if icon, exists := icons[field]; exists {
		return icon
	}
	return "📝"
}

// escapeMarkdownV2 экранирует текст по спецификации MarkdownV2 для Telegram
func escapeMarkdownV2(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// escapeMarkdownV2LinkText экранирует текст ссылки по спецификации MarkdownV2 для Telegram
func escapeMarkdownV2LinkText(text string) string {
	specialChars := []string{"_", "*", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// escapeMarkdownV2URL экранирует URL ссылки по спецификации MarkdownV2 для Telegram
func escapeMarkdownV2URL(url string) string {
	result := url
	result = strings.ReplaceAll(result, "\\", "\\\\")
	result = strings.ReplaceAll(result, ")", "\\)")
	return result
}
