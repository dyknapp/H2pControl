package templates

import "embed"

// If you do not set this to all:server, files/dirs starting with _ or . are not included:
// https://pkg.go.dev/embed#hdr-Directives

//go:embed all:*
var TemplatesFS embed.FS
