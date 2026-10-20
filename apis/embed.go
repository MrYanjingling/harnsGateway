// Package apidocs embeds the OpenAPI specifications served by the platform.
package apidocs

import "embed"

//go:embed platform/*.yaml
var FS embed.FS
