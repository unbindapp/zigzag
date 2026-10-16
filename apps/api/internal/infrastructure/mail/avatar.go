package mail

import (
	"strings"
	"unicode/utf16"
)

var avatarPalette = []string{
	"#95a2b3",
	"#5e6ad2",
	"#26b5ce",
	"#0f783c",
	"#4cb782",
	"#f2c94c",
	"#f2994a",
	"#eb5757",
	"#db6e9e",
	"#bb87fc",
	"#8e6b4a",
	"#4b5563",
}

func AvatarColor(id string) string {
	var hash int32
	for _, unit := range utf16.Encode([]rune(id)) {
		hash = hash*31 + int32(unit)
	}
	magnitude := int64(hash)
	if magnitude < 0 {
		magnitude = -magnitude
	}
	return avatarPalette[magnitude%int64(len(avatarPalette))]
}

func Initials(name string) string {
	words := strings.Fields(name)
	switch len(words) {
	case 0:
		return "?"
	case 1:
		return strings.ToUpper(string(firstRunes(words[0], 2)))
	default:
		return strings.ToUpper(string(firstRunes(words[0], 1)) + string(firstRunes(words[len(words)-1], 1)))
	}
}

func firstRunes(word string, count int) []rune {
	runes := []rune(word)
	return runes[:min(count, len(runes))]
}
