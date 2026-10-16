package mail

import (
	"strings"
	"testing"
)

func TestRenderReferencesInlineLogo(t *testing.T) {
	msg, err := Render("to@example.com", Content{Subject: "s", Heading: "h"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg.HTML, `src="cid:logo.png"`) {
		t.Fatalf("logo not referenced by cid: %s", msg.HTML)
	}
	if len(msg.Inline) != 1 || msg.Inline[0].Name != "logo.png" {
		t.Fatalf("logo not embedded: %+v", msg.Inline)
	}
}
