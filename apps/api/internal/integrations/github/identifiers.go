package github

import (
	"regexp"
	"strconv"
	"strings"
)

const MaxReferences = 50

type Reference struct {
	Key    string
	Number int
	Closes bool
}

func (self Reference) Identifier() string {
	return self.Key + "-" + strconv.Itoa(self.Number)
}

const identifier = `([a-z][a-z0-9]{0,6})-([0-9]{1,9})`

var (
	branchIdentifier  = regexp.MustCompile(`(?i)(?:^|[/_-])` + identifier + `(?:$|[^a-z0-9])`)
	textIdentifier    = regexp.MustCompile(`(?i)\b` + identifier + `\b`)
	closingPhrase     = regexp.MustCompile(`(?i)\b(?:fix(?:es|ed)?|close[sd]?|resolve[sd]?)\b:?((?:(?:\s*,\s*|\s+and\s+|\s+)\[?[a-z][a-z0-9]{0,6}-[0-9]{1,9}\]?)+)`)
	closingIdentifier = regexp.MustCompile(`(?i)` + identifier)
)

func References(branch string, texts ...string) []Reference {
	var found []Reference
	add := func(match []string, closes bool) {
		number, err := strconv.Atoi(match[2])
		if err != nil || number <= 0 {
			return
		}
		ref := Reference{Key: strings.ToUpper(match[1]), Number: number, Closes: closes}
		for i := range found {
			if found[i].Key == ref.Key && found[i].Number == ref.Number {
				found[i].Closes = found[i].Closes || closes
				return
			}
		}
		if len(found) < MaxReferences {
			found = append(found, ref)
		}
	}
	for _, match := range findOverlapping(branchIdentifier, branch) {
		add(match, false)
	}
	for _, text := range texts {
		for _, phrase := range closingPhrase.FindAllStringSubmatch(text, -1) {
			for _, match := range closingIdentifier.FindAllStringSubmatch(phrase[1], -1) {
				add(match, true)
			}
		}
		for _, match := range textIdentifier.FindAllStringSubmatch(text, -1) {
			add(match, false)
		}
	}
	return found
}

func findOverlapping(pattern *regexp.Regexp, text string) [][]string {
	var matches [][]string
	for start := 0; start < len(text); {
		loc := pattern.FindStringSubmatchIndex(text[start:])
		if loc == nil {
			break
		}
		matches = append(matches, []string{text[start+loc[0] : start+loc[1]], text[start+loc[2] : start+loc[3]], text[start+loc[4] : start+loc[5]]})
		start += loc[5]
	}
	return matches
}
