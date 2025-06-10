package core

import (
	"strings"

	"a-la-carte/internal/app"

	"github.com/mattn/go-runewidth"
)

// emojiRule defines a mapping from keywords to an emoji.
type emojiRule struct {
	matches []string
	emoji   string
}

// emojiRules is the list of rules for matching software entries to emojis.
var emojiRules = []emojiRule{
	{matches: []string{"python"}, emoji: "🐍"},
	{matches: []string{"node", "node.js"}, emoji: "🟩"},
	{matches: []string{"go", "golang"}, emoji: "🐹"},
	{matches: []string{"docker"}, emoji: "🐳"},
	{matches: []string{"git"}, emoji: "🌱"},
	{matches: []string{"linux"}, emoji: "🐧"},
	{matches: []string{"mac", "apple"}, emoji: "🍏"},
	{matches: []string{"brew"}, emoji: "🍺"},
	{matches: []string{"terminal", "cli", "tui"}, emoji: "💻"},
	{matches: []string{"test", "testing"}, emoji: "🧪"},
	{matches: []string{"file", "document"}, emoji: "📄"},
	{matches: []string{"key", "password", "secret"}, emoji: "🔑"},
	{matches: []string{"sync", "update"}, emoji: "🔄"},
	{matches: []string{"note", "write"}, emoji: "📝"},
	{matches: []string{"package", "install"}, emoji: "📦"},
	{matches: []string{"tool", "utility"}, emoji: "🧰"},
}

// checkContains returns true if any of the matches are found in name or desc.
func checkContains(name, desc string, matches ...string) bool {
	n := strings.ToLower(name)
	d := strings.ToLower(desc)
	for _, m := range matches {
		if strings.Contains(n, m) || strings.Contains(d, m) {
			return true
		}
	}
	return false
}

// NormalizeEmoji ensures the emoji is exactly 2 columns wide for consistent alignment.
//
// # Parameters
//   - e: the emoji string
//
// # Returns
//   - The normalized emoji string, always 2 columns wide.
func NormalizeEmoji(e string) string {
	w := runewidth.StringWidth(e)
	if w == 2 {
		return e
	}
	if w > 2 {
		runes := []rune(e)
		acc := 0
		for i, r := range runes {
			acc += runewidth.RuneWidth(r)
			if acc >= 2 {
				return string(runes[:i+1])
			}
		}
		return string(runes[:1]) + " "
	}
	// pad with space if too narrow
	return e + strings.Repeat(" ", 2-w)
}

// EmojiForEntry returns the best-matching emoji for a software entry.
//
// # Parameters
//   - e: pointer to the SoftwareEntry
//
// # Returns
//   - The emoji string, always 2 columns wide.
func EmojiForEntry(e *app.SoftwareEntry) string {
	for _, rule := range emojiRules {
		if checkContains(e.Name, e.Desc, rule.matches...) {
			return NormalizeEmoji(rule.emoji)
		}
	}
	return NormalizeEmoji("📦") // default emoji
}
