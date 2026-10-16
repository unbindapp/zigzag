package mail

import (
	"bytes"
	"embed"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"
)

//go:embed templates/*
var templateFS embed.FS

const (
	logoName            = "logo.png"
	defaultInstanceName = "Zigzag"
)

var layoutFuncs = htmltemplate.FuncMap{
	"avatarColor":     AvatarColor,
	"msoStyle":        constHTML(`<!--[if mso]><style>table,td,a,span{font-family:'Segoe UI',Arial,sans-serif !important;}</style><![endif]-->`),
	"msoOpen":         constHTML(`<!--[if mso]><table role="presentation" width="600" align="center" cellpadding="0" cellspacing="0" border="0"><tr><td><![endif]-->`),
	"msoClose":        constHTML(`<!--[if mso]></td></tr></table><![endif]-->`),
	"preheaderSpacer": constHTML(strings.Repeat("&zwnj;&nbsp;", 90)),
}

func constHTML(markup string) func() htmltemplate.HTML {
	return func() htmltemplate.HTML { return htmltemplate.HTML(markup) }
}

var (
	htmlTemplate = htmltemplate.Must(htmltemplate.New("layout.html").Funcs(layoutFuncs).ParseFS(templateFS, "templates/layout.html"))
	textTemplate = texttemplate.Must(texttemplate.ParseFS(templateFS, "templates/layout.txt"))
	logoPNG      = mustRead("templates/" + logoName)
)

func mustRead(name string) []byte {
	data, err := templateFS.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return data
}

type Content struct {
	Subject   string
	Preheader string
	Heading   string
	Intro     []string
	Items     []Item
	Button    Button
	Note      string
	Footer    Footer
}

type Item struct {
	Identifier string
	Title      string
	URL        string
	Lines      []Line
}

type Line struct {
	ActorID       string
	ActorName     string
	ActorInitials string
	Text          string
}

type Button struct {
	Label string
	URL   string
}

type Footer struct {
	InstanceName string
	InstanceURL  string
	ManageURL    string
	ManageLabel  string
}

type layout struct {
	Content
	LogoSrc htmltemplate.URL
}

func ActorLine(actorID, actorName, text string) Line {
	return Line{ActorID: actorID, ActorName: actorName, ActorInitials: Initials(actorName), Text: text}
}

func Render(to string, content Content) (Message, error) {
	if content.Footer.InstanceName == "" {
		content.Footer.InstanceName = defaultInstanceName
	}
	if content.Preheader == "" && len(content.Intro) > 0 {
		content.Preheader = content.Intro[0]
	}
	var text, html bytes.Buffer
	if err := textTemplate.Execute(&text, content); err != nil {
		return Message{}, err
	}
	if err := htmlTemplate.Execute(&html, layout{Content: content, LogoSrc: htmltemplate.URL("cid:" + logoName)}); err != nil {
		return Message{}, err
	}
	return Message{
		To:      to,
		Subject: content.Subject,
		Text:    text.String(),
		HTML:    html.String(),
		Inline:  []Inline{{Name: logoName, ContentType: "image/png", Data: logoPNG}},
	}, nil
}
