package tui

import "fmt"

func pluralWord(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func preservedMessagesPhrase(count int) string {
	return fmt.Sprintf("%d %s kept", count, pluralWord(count, "message", "messages"))
}