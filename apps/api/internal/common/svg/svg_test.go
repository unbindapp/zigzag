package svg

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
)

func TestValidateAcceptsSafeImages(t *testing.T) {
	favicon, err := os.ReadFile("../../../../web/public/favicon.svg")
	require.NoError(t, err)

	tests := []struct {
		name   string
		input  string
		width  int
		height int
	}{
		{name: "repo favicon", input: string(favicon), width: 1254, height: 1254},
		{name: "plain icon from viewBox", input: `<?xml version="1.0" encoding="UTF-8"?><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M12 2L2 22h20z"/></svg>`, width: 24, height: 24},
		{name: "style block", input: `<svg xmlns="http://www.w3.org/2000/svg" width="48px" height="32"><style>.a{fill:#f00}</style><rect class="a" width="48" height="32"/></svg>`, width: 48, height: 32},
		{name: "width and height win over viewBox", input: `<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 16 16"/>`, width: 64, height: 64},
		{name: "fractional viewBox rounds", input: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0,0,10.6,20.4"/>`, width: 11, height: 20},
		{name: "relative sizes fall back to viewBox", input: `<svg xmlns="http://www.w3.org/2000/svg" width="100%" height="100%" viewBox="0 0 30 10"/>`, width: 30, height: 10},
		{name: "no size", input: `<svg xmlns="http://www.w3.org/2000/svg" width="2em" height="2em"><circle r="4"/></svg>`},
		{name: "internal use and data image", input: `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 8 8"><defs><circle id="c" r="2"/></defs><use xlink:href="#c"/><image href="data:image/png;base64,iVBORw0KGgo="/></svg>`, width: 8, height: 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height, err := Validate([]byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.width, width)
			assert.Equal(t, tt.height, height)
		})
	}
}

func TestValidateRejectsUnsafeImages(t *testing.T) {
	const open = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">`
	tests := []struct {
		name  string
		input string
	}{
		{name: "not xml", input: `<svg><path></svg>`},
		{name: "empty", input: ``},
		{name: "root is not svg", input: `<html><svg/></html>`},
		{name: "script element", input: open + `<script>alert(1)</script></svg>`},
		{name: "namespaced html script", input: open + `<h:script xmlns:h="http://www.w3.org/1999/xhtml">alert(1)</h:script></svg>`},
		{name: "upper case script", input: open + `<SCRIPT>alert(1)</SCRIPT></svg>`},
		{name: "onload attribute", input: `<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"/>`},
		{name: "onclick on a child", input: open + `<rect onClick="alert(1)"/></svg>`},
		{name: "javascript href", input: open + `<a href="javascript:alert(1)"><text>x</text></a></svg>`},
		{name: "javascript xlink href with whitespace and entities", input: open + `<a xlink:href=" java&#x09;script&#58;alert(1)"><text>x</text></a></svg>`},
		{name: "javascript in style", input: open + `<rect style="background:url(javascript:alert(1))"/></svg>`},
		{name: "html data url in src", input: open + `<image src="data:text/html;base64,PHNjcmlwdD4="/></svg>`},
		{name: "animated href", input: open + `<a><animate attributeName="href" values="javascript:alert(1)"/><text>x</text></a></svg>`},
		{name: "foreignObject", input: open + `<foreignObject><div xmlns="http://www.w3.org/1999/xhtml">hi</div></foreignObject></svg>`},
		{name: "doctype", input: `<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">` + open + `</svg>`},
		{name: "external entity", input: `<!DOCTYPE svg [<!ENTITY x SYSTEM "file:///etc/passwd">]>` + open + `<text>&x;</text></svg>`},
		{name: "undeclared entity", input: open + `<text>&x;</text></svg>`},
		{name: "stylesheet instruction", input: `<?xml-stylesheet href="https://example.com/a.css"?>` + open + `</svg>`},
		{name: "remote image", input: open + `<image href="https://example.com/a.png"/></svg>`},
		{name: "remote image via xlink", input: open + `<image xlink:href="HTTP://example.com/a.png"/></svg>`},
		{name: "remote use", input: open + `<use href="http://example.com/sprite.svg#a"/></svg>`},
		{name: "protocol relative use", input: open + `<use href="//example.com/sprite.svg#a"/></svg>`},
		{name: "too large", input: open + strings.Repeat(" ", MaxBytes) + `</svg>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := Validate([]byte(tt.input))
			assert.ErrorIs(t, err, errdefs.ErrInvalidInput)
		})
	}
}
