// Package web embeds the built frontend of the local service: web/dist is
// committed, so `go install` and GoReleaser build the binary without Node.
// `make web-check` keeps dist in step with the sources.
package web

import (
	"embed"
	"io/fs"
)

//go:embed dist
var dist embed.FS

// Dist returns the built frontend: index.html and assets/ at its root.
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
