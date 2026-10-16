package mail

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var footer = Footer{InstanceName: "Zigzag", InstanceURL: "https://zigzag.example.com"}

func digestContent() Content {
	return Content{
		Subject:   "3 unread notifications",
		Preheader: "ngyekta assigned the issue to you",
		Heading:   "You have 3 unread notifications in Acme.",
		Items: []Item{
			{
				Identifier: "ENG-12",
				Title:      "Fix login",
				URL:        "https://zigzag.example.com/acme/issue/ENG-12",
				Lines: []Line{
					ActorLine("0b9f6c1e-5d1a-4c55-9a43-3f7f1a2b8c11", "ngyekta", "assigned the issue to you"),
					ActorLine("7d4e2f60-1c2b-4a8e-b0a1-9e5c3d2f4a77", "Ada Lovelace", "changed the status to Done"),
				},
			},
			{
				Identifier: "ENG-7",
				Title:      "Billing page",
				URL:        "https://zigzag.example.com/acme/issue/ENG-7",
				Lines:      []Line{{Text: "Due today"}},
			},
		},
		Button: Button{Label: "Open your inbox", URL: "https://zigzag.example.com/acme/inbox"},
		Footer: Footer{
			InstanceName: "Zigzag",
			InstanceURL:  "https://zigzag.example.com",
			ManageURL:    "https://zigzag.example.com/acme/settings/account/notifications",
			ManageLabel:  "Notification settings",
		},
	}
}

func TestRenderDigest(t *testing.T) {
	msg, err := Render("grace@example.com", digestContent())
	require.NoError(t, err)

	assert.Equal(t, "grace@example.com", msg.To)
	assert.Equal(t, "3 unread notifications", msg.Subject)
	assert.Equal(t, `You have 3 unread notifications in Acme.

ENG-12 Fix login
    ngyekta assigned the issue to you
    Ada Lovelace changed the status to Done
    https://zigzag.example.com/acme/issue/ENG-12

ENG-7 Billing page
    Due today
    https://zigzag.example.com/acme/issue/ENG-7

Open your inbox
https://zigzag.example.com/acme/inbox

--
Zigzag https://zigzag.example.com
Notification settings: https://zigzag.example.com/acme/settings/account/notifications
`, msg.Text)

	for _, block := range []string{
		`<meta name="color-scheme" content="light dark">`,
		`@media (prefers-color-scheme: dark)`,
		`>ngyekta assigned the issue to you`,
		`<a href="https://zigzag.example.com/acme/issue/ENG-12"`,
		`>ENG-12</span>`,
		`Fix login</a>`,
		`bgcolor="#f2994a"`,
		`background-color:#f2994a;`,
		`>NG</td>`,
		`>AL</td>`,
		`<span class="zz-strong" style="font-weight:600;color:#0f131a;">Ada Lovelace</span> changed the status to Done`,
		`>Due today</td>`,
		`<a href="https://zigzag.example.com/acme/inbox"`,
		`>Open your inbox</a>`,
		`<a href="https://zigzag.example.com/acme/settings/account/notifications" class="zz-muted" style="color:#606876;text-decoration:underline;">Notification settings</a>`,
		`<!--[if mso]><table role="presentation" width="600"`,
	} {
		assert.Contains(t, msg.HTML, block)
	}
	assert.NotContains(t, msg.HTML, "Or open this link", "digest items are links, the fallback url is left out")
	assert.Equal(t, 2, strings.Count(msg.HTML, `width="20" height="20" align="center"`), "one avatar per actor line")
}

func TestRenderInvitation(t *testing.T) {
	msg, err := Render("new@example.com", Content{
		Subject: "Ada Lovelace invited you to Acme on Zigzag",
		Heading: "Join Acme on Zigzag",
		Intro:   []string{"Ada Lovelace invited you to the Acme workspace as a member.", "Members create and edit issues, projects and documents."},
		Button:  Button{Label: "Accept invitation", URL: "https://zigzag.example.com/invite/abc"},
		Note:    "This invitation expires in 7 days. If you were not expecting it, you can ignore this email.",
		Footer:  footer,
	})
	require.NoError(t, err)

	assert.Equal(t, `Join Acme on Zigzag

Ada Lovelace invited you to the Acme workspace as a member.

Members create and edit issues, projects and documents.

Accept invitation
https://zigzag.example.com/invite/abc

This invitation expires in 7 days. If you were not expecting it, you can ignore this email.

--
Zigzag https://zigzag.example.com
`, msg.Text)
	assert.Contains(t, msg.HTML, ">Join Acme on Zigzag</td>")
	assert.Contains(t, msg.HTML, `mso-hide:all;">Ada Lovelace invited you to the Acme workspace as a member.&zwnj;`, "the first paragraph is the preheader")
	assert.Contains(t, msg.HTML, `<td class="zz-button" bgcolor="#4753d1"`)
	assert.Contains(t, msg.HTML, `Or open this link: <a href="https://zigzag.example.com/invite/abc"`)
	assert.Contains(t, msg.HTML, ">This invitation expires in 7 days.")
	assert.NotContains(t, msg.HTML, "Notification settings")
	assert.NotContains(t, msg.HTML, `class="zz-rule"`, "no item list without items")
}

func TestRenderPreheaderDefaultsToTheFirstParagraph(t *testing.T) {
	msg, err := Render("ada@example.com", Content{Subject: "Reset", Heading: "Reset your password", Intro: []string{"Someone asked to reset the password for ada@example.com."}})
	require.NoError(t, err)
	assert.Contains(t, msg.HTML, `mso-hide:all;">Someone asked to reset the password for ada@example.com.&zwnj;`)
}

func TestRenderPasswordReset(t *testing.T) {
	msg, err := Render("ada@example.com", Content{
		Subject: "Reset your Zigzag password",
		Heading: "Reset your password",
		Intro:   []string{"Someone asked to reset the password for ada@example.com. Choose a new one with the button below."},
		Button:  Button{Label: "Reset password", URL: "https://zigzag.example.com/reset-password?token=t0k&x=1"},
		Note:    "The link works once and expires in one hour. If you did not ask for it, ignore this email and your password stays the same.",
		Footer:  footer,
	})
	require.NoError(t, err)

	assert.Contains(t, msg.Text, "Reset password\nhttps://zigzag.example.com/reset-password?token=t0k&x=1\n")
	assert.Contains(t, msg.HTML, `href="https://zigzag.example.com/reset-password?token=t0k&amp;x=1"`)
	assert.Contains(t, msg.HTML, ">Reset password</a>")
	assert.Contains(t, msg.HTML, "If you did not ask for it, ignore this email")
}

func TestRenderTestEmail(t *testing.T) {
	msg, err := Render("ada@example.com", Content{
		Subject: "Zigzag can send email",
		Heading: "Email is working",
		Intro:   []string{"This test message came from your Zigzag instance, so its mail settings work."},
		Footer:  Footer{InstanceName: "Acme Tracker"},
	})
	require.NoError(t, err)

	assert.Equal(t, "Email is working\n\nThis test message came from your Zigzag instance, so its mail settings work.\n\n--\nAcme Tracker\n", msg.Text)
	assert.Contains(t, msg.HTML, ">Acme Tracker</td>")
	assert.NotContains(t, msg.HTML, `href=""`, "no button and no instance link without urls")
	assert.NotContains(t, msg.HTML, `class="zz-button"`)
}

func TestRenderDefaultsTheInstanceName(t *testing.T) {
	msg, err := Render("ada@example.com", Content{Subject: "s", Heading: "h"})
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(msg.Text, "--\nZigzag\n"))
}

func TestRenderEscapesUserInput(t *testing.T) {
	content := digestContent()
	content.Heading = "Join <b>Acme</b>"
	content.Intro = []string{`<img src=x onerror="alert(1)">`}
	content.Items[0].Title = "<script>alert(1)</script>"
	content.Items[0].URL = "javascript:alert(1)"
	content.Items[0].Lines[0].ActorName = `<a href="https://evil.example">Mallory</a>`
	content.Items[0].Lines[0].ActorID = `"><script>alert(1)</script>`
	content.Footer.InstanceName = "<i>Evil</i>"

	msg, err := Render("grace@example.com", content)
	require.NoError(t, err)

	for _, raw := range []string{"<b>Acme", "<img src=x", "<script>", "javascript:alert", `<a href="https://evil.example">`, "<i>Evil", `"><script>`} {
		assert.NotContains(t, msg.HTML, raw)
	}
	assert.Contains(t, msg.HTML, "Join &lt;b&gt;Acme&lt;/b&gt;")
	assert.Contains(t, msg.HTML, "&lt;script&gt;alert(1)&lt;/script&gt;")
	assert.Contains(t, msg.HTML, `href="#ZgotmplZ"`)
	assert.Contains(t, msg.Text, "<script>alert(1)</script>", "the text part stays verbatim")
}

func TestAvatarMatchesTheWebApp(t *testing.T) {
	for id, color := range map[string]string{
		"0b9f6c1e-5d1a-4c55-9a43-3f7f1a2b8c11": "#f2994a",
		"7d4e2f60-1c2b-4a8e-b0a1-9e5c3d2f4a77": "#26b5ce",
		"c2a1e8f4-9b3d-4f6a-8e2c-1d5b7a9c3e05": "#95a2b3",
		"":                                     "#95a2b3",
		"a":                                    "#5e6ad2",
	} {
		assert.Equal(t, color, AvatarColor(id), id)
	}
	for name, initials := range map[string]string{
		"Ada Lovelace":               "AL",
		"ngyekta":                    "NG",
		"  grace  brewster  hopper ": "GH",
		"":                           "?",
		"émile zola":                 "ÉZ",
	} {
		assert.Equal(t, initials, Initials(name), name)
	}
}
