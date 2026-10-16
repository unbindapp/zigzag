package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ref(identifier string, number int, closes bool) Reference {
	return Reference{Key: identifier, Number: number, Closes: closes}
}

func TestReferences(t *testing.T) {
	cases := []struct {
		name   string
		branch string
		texts  []string
		want   []Reference
	}{
		{name: "user prefixed branch with slug", branch: "bbedward/unb-123-slug", want: []Reference{ref("UNB", 123, false)}},
		{name: "bare branch", branch: "unb-123", want: []Reference{ref("UNB", 123, false)}},
		{name: "uppercase branch below a folder", branch: "feature/UNB-123", want: []Reference{ref("UNB", 123, false)}},
		{name: "two identifiers in one branch", branch: "unb-1-unb-2", want: []Reference{ref("UNB", 1, false), ref("UNB", 2, false)}},
		{name: "branch without an identifier", branch: "main"},
		{name: "identifier glued to a word in a branch", branch: "xunb-12", want: []Reference{ref("XUNB", 12, false)}},
		{name: "identifier followed by letters in a branch", branch: "unb-12abc"},
		{name: "title mention", texts: []string{"UNB-123 tighten the websocket limit"}, want: []Reference{ref("UNB", 123, false)}},
		{name: "bracketed title mention", texts: []string{"[UNB-123] tighten the websocket limit"}, want: []Reference{ref("UNB", 123, false)}},
		{name: "lowercase mention", texts: []string{"see unb-7"}, want: []Reference{ref("UNB", 7, false)}},
		{name: "magic word", texts: []string{"fixes UNB-123"}, want: []Reference{ref("UNB", 123, true)}},
		{name: "magic word with a list", texts: []string{"Fixes UNB-123, UNB-124"}, want: []Reference{ref("UNB", 123, true), ref("UNB", 124, true)}},
		{name: "magic word with and", texts: []string{"Resolves UNB-1 and UNB-2"}, want: []Reference{ref("UNB", 1, true), ref("UNB", 2, true)}},
		{name: "magic word with a colon and brackets", texts: []string{"Closes: [UNB-9]"}, want: []Reference{ref("UNB", 9, true)}},
		{name: "every magic word", texts: []string{"fix A-1 fixed A-2 close A-3 closed A-4 resolve A-5 resolved A-6 CLOSES A-7"}, want: []Reference{
			ref("A", 1, true), ref("A", 2, true), ref("A", 3, true), ref("A", 4, true), ref("A", 5, true), ref("A", 6, true), ref("A", 7, true),
		}},
		{name: "magic word list stops at prose", texts: []string{"Fixes UNB-1 and also touches UNB-2"}, want: []Reference{ref("UNB", 1, true), ref("UNB", 2, false)}},
		{name: "word ending in a magic word", texts: []string{"enclosed UNB-3"}, want: []Reference{ref("UNB", 3, false)}},
		{name: "magic word without an identifier", texts: []string{"fixes the flaky test"}},
		{name: "github issue numbers are not identifiers", texts: []string{"Fixes #12"}},
		{name: "zero is not an issue number", texts: []string{"UNB-0"}},
		{name: "key longer than seven characters", texts: []string{"ABCDEFGH-1"}},
		{name: "branch and closing body merge into one reference", branch: "bbedward/unb-455-limit", texts: []string{"Limit reads", "fixes UNB-455"}, want: []Reference{ref("UNB", 455, true)}},
		{name: "commit message with a trailer on the last line", texts: []string{"Limit websocket reads\n\nThe reader caps frames, see UNB-4.\n\nFixes UNB-5"}, want: []Reference{ref("UNB", 5, true), ref("UNB", 4, false)}},
		{name: "commit message with trailers after the closing one", texts: []string{"Cap frames\n\nCloses UNB-5, UNB-6\nCo-authored-by: Brandon <b@example.com>\n"}, want: []Reference{ref("UNB", 5, true), ref("UNB", 6, true)}},
		{name: "closing list stops at the end of its line", texts: []string{"Fixes UNB-5\nUNB-6 follows up"}, want: []Reference{ref("UNB", 5, true), ref("UNB", 6, false)}},
		{name: "magic word on a later line of a commit message", texts: []string{"Refactor the reader for UNB-7\n\nresolved: UNB-8"}, want: []Reference{ref("UNB", 8, true), ref("UNB", 7, false)}},
		{name: "mention then magic word keeps the first position", texts: []string{"UNB-2 then UNB-1", "closes unb-2"}, want: []Reference{ref("UNB", 2, true), ref("UNB", 1, false)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, References(tc.branch, tc.texts...))
		})
	}
}

func TestReferencesAreCapped(t *testing.T) {
	text := ""
	for i := 1; i <= MaxReferences+10; i++ {
		text += " UNB-" + string(rune('0'+i%10)) + string(rune('0'+i/10))
	}
	assert.Len(t, References("", text), MaxReferences)
}

func TestReferenceIdentifier(t *testing.T) {
	assert.Equal(t, "UNB-455", ref("UNB", 455, false).Identifier())
}
