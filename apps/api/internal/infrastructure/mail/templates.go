package mail

import (
	"bytes"
	"embed"
	htmltemplate "html/template"
	texttemplate "text/template"
)

//go:embed templates/*
var templateFS embed.FS

var (
	htmlTemplate = htmltemplate.Must(htmltemplate.ParseFS(templateFS, "templates/layout.html"))
	textTemplate = texttemplate.Must(texttemplate.ParseFS(templateFS, "templates/layout.txt"))
)

type Content struct {
	Subject      string
	Heading      string
	Paragraphs   []string
	ButtonLabel  string
	ButtonURL    string
	Note         string
	InstanceName string
	InstanceURL  string
}

func Render(to string, content Content) (Message, error) {
	if content.InstanceName == "" {
		content.InstanceName = "Zigzag"
	}
	var text, html bytes.Buffer
	if err := textTemplate.Execute(&text, content); err != nil {
		return Message{}, err
	}
	if err := htmlTemplate.Execute(&html, content); err != nil {
		return Message{}, err
	}
	return Message{
		To:      to,
		Subject: content.Subject,
		Text:    text.String(),
		HTML:    html.String(),
	}, nil
}
