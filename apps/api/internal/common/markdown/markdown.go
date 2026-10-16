package markdown

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	mentionPattern     = regexp.MustCompile(`@\[([^\]]*)\]\(user:([0-9a-fA-F-]{36})\)`)
	imagePattern       = regexp.MustCompile(`!\[([^\]]*)\]\([^)]*\)`)
	linkPattern        = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	referenceLink      = regexp.MustCompile(`(?m)^\s*\[[^\]]+\]:\s+\S+.*$`)
	fencePattern       = regexp.MustCompile("(?m)^[ \\t]*(```|~~~).*(\\n|$)")
	inlineCode         = regexp.MustCompile("`([^`]*)`")
	htmlTag            = regexp.MustCompile(`</?[a-zA-Z][^>]*>`)
	headingMarker      = regexp.MustCompile(`(?m)^\s{0,3}#{1,6}\s+`)
	quoteMarker        = regexp.MustCompile(`(?m)^\s*(>\s?)+`)
	listMarker         = regexp.MustCompile(`(?m)^\s*([-*+]|\d+[.)])\s+(\[[ xX]\]\s+)?`)
	ruleLine           = regexp.MustCompile(`(?m)^\s*([-*_]\s*){3,}$`)
	starEmphasis       = regexp.MustCompile(`(\*\*|~~|\*)([^\s*~](?:.*?[^\s*~])?)(\*\*|~~|\*)`)
	underscoreEmphasis = regexp.MustCompile(`(^|[^\w])(__|_)([^\s_](?:.*?[^\s_])?)(__|_)($|[^\w])`)
	tableSeparatorLine = regexp.MustCompile(`(?m)^\s*\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)*\|?\s*$`)
	whitespace         = regexp.MustCompile(`[ \t]+`)
	blankLines         = regexp.MustCompile(`\n{3,}`)
)

func PlainText(md string) string {
	text := strings.ReplaceAll(md, "\r\n", "\n")
	text = mentionPattern.ReplaceAllString(text, "@$1")
	text = fencePattern.ReplaceAllString(text, "")
	text = imagePattern.ReplaceAllString(text, "$1")
	text = linkPattern.ReplaceAllString(text, "$1")
	text = referenceLink.ReplaceAllString(text, "")
	text = inlineCode.ReplaceAllString(text, "$1")
	text = htmlTag.ReplaceAllString(text, "")
	text = ruleLine.ReplaceAllString(text, "")
	text = tableSeparatorLine.ReplaceAllString(text, "")
	text = headingMarker.ReplaceAllString(text, "")
	text = quoteMarker.ReplaceAllString(text, "")
	text = listMarker.ReplaceAllString(text, "")
	for range 3 {
		text = starEmphasis.ReplaceAllString(text, "$2")
		text = underscoreEmphasis.ReplaceAllString(text, "$1$3$5")
	}
	text = strings.ReplaceAll(text, "|", " ")
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(whitespace.ReplaceAllString(line, " "))
	}
	text = strings.Join(lines, "\n")
	text = blankLines.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

func Mentions(md string) []uuid.UUID {
	matches := mentionPattern.FindAllStringSubmatch(md, -1)
	seen := map[uuid.UUID]bool{}
	ids := []uuid.UUID{}
	for _, match := range matches {
		id, err := uuid.Parse(match[2])
		if err != nil || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids
}
