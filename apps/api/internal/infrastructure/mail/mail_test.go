package mail

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderBothParts(t *testing.T) {
	msg, err := Render("ada@example.com", Content{
		Subject:     "Join Acme",
		Heading:     "Join <Acme>",
		Paragraphs:  []string{"Bob invited you.", "Accept to get started."},
		ButtonLabel: "Accept invitation",
		ButtonURL:   "https://zigzag.example.com/invite/abc",
		Note:        "Expires in 7 days.",
		InstanceURL: "https://zigzag.example.com",
	})
	require.NoError(t, err)

	assert.Equal(t, "ada@example.com", msg.To)
	assert.Equal(t, "Join Acme", msg.Subject)
	for _, part := range []string{msg.Text, msg.HTML} {
		assert.Contains(t, part, "Bob invited you.")
		assert.Contains(t, part, "https://zigzag.example.com/invite/abc")
		assert.Contains(t, part, "Expires in 7 days.")
		assert.Contains(t, part, "Sent by Zigzag")
	}
	assert.Contains(t, msg.HTML, "Join &lt;Acme&gt;", "html escapes content")
	assert.Contains(t, msg.Text, "Accept invitation: https://zigzag.example.com/invite/abc")
}

func TestRenderWithoutButton(t *testing.T) {
	msg, err := Render("a@example.com", Content{Subject: "Test", Heading: "Email is working", InstanceName: "Acme Tracker"})
	require.NoError(t, err)
	assert.False(t, strings.Contains(msg.HTML, "<a href=\"\""))
	assert.Contains(t, msg.HTML, "Acme Tracker")
}
