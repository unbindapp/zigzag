package markdown

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPlainText(t *testing.T) {
	tests := map[string]struct {
		in   string
		want string
	}{
		"empty":         {"", ""},
		"plain":         {"just words", "just words"},
		"heading":       {"# Title\n\nBody", "Title\n\nBody"},
		"emphasis":      {"some **bold**, _italic_ and ~~gone~~ text", "some bold, italic and gone text"},
		"links":         {"see [the docs](https://example.com) now", "see the docs now"},
		"images":        {"![diagram](/api/go/uploads/get?upload_id=1) caption", "diagram caption"},
		"inline code":   {"run `make check` first", "run make check first"},
		"code fence":    {"before\n```go\nfmt.Println(1)\n```\nafter", "before\nfmt.Println(1)\nafter"},
		"lists":         {"- one\n* two\n1. three\n- [x] done", "one\ntwo\nthree\ndone"},
		"quote":         {"> quoted\n> lines", "quoted\nlines"},
		"html":          {"a <b>bold</b> tag", "a bold tag"},
		"mention":       {"ping @[Ada Lovelace](user:7f9c4d2b-0000-4000-8000-000000000042) please", "ping @Ada Lovelace please"},
		"rule":          {"above\n\n---\n\nbelow", "above\n\nbelow"},
		"crlf":          {"one\r\ntwo", "one\ntwo"},
		"blank lines":   {"a\n\n\n\n\nb", "a\n\nb"},
		"table":         {"| a | b |\n| --- | --- |\n| 1 | 2 |", "a b\n\n1 2"},
		"snake_case id": {"call get_issue_by_id", "call get_issue_by_id"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, PlainText(tt.in))
		})
	}
}

func TestMentions(t *testing.T) {
	a := uuid.MustParse("7f9c4d2b-0000-4000-8000-000000000042")
	b := uuid.MustParse("11111111-2222-4333-8444-555555555555")

	assert.Equal(t, []uuid.UUID{}, Mentions("no mentions here"))
	assert.Equal(t, []uuid.UUID{a}, Mentions("@[Ada](user:"+a.String()+")"))
	assert.Equal(t, []uuid.UUID{a, b}, Mentions("@[Ada](user:"+a.String()+") and @[Bob](user:"+b.String()+") and again @[Ada](user:"+a.String()+")"))
	assert.Equal(t, []uuid.UUID{}, Mentions("@[Broken](user:not-a-uuid-at-all-not-a-uuid-at-all-12)"))
	assert.Equal(t, []uuid.UUID{}, Mentions("[Ada](user:"+a.String()+") without the at sign"))
}
