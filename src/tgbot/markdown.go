package tgbot

import (
	"fmt"

	api "github.com/go-telegram-bot-api/telegram-bot-api"
)

// BordText transforms text into markdown bord text.
func BordText(text string) string {
	return fmt.Sprintf("*%s*", text)
}

// InlineLink combines text and link into markdown inline link.
func InlineLink(text, link string) string {
	return fmt.Sprintf("[%s](%s)", text, link)
}

// ItalicText transforms text into markdown italic text.
func ItalicText(text string) string {
	return fmt.Sprintf("_%s_", text)
}

// Inlice code transforms text into markdown inline code.
func InlineCode(text string) string {
	return fmt.Sprintf("`%s`", text)
}

// EscapeText takes an input text and escape Telegram markup symbols.
// In this way we can send a text without being afraid of having to escape the characters manually.
// Note that you don't have to include the formatting style in the input text, or it will be escaped too.
func EscapeText(text string) string {
	return api.EscapeText(api.ModeMarkdownV2, text)
}
