package web

import "embed"

// the committed dist/index.html is a stub; make embed replaces dist/ with the real SPA build
//
//go:embed all:dist
var distFS embed.FS
