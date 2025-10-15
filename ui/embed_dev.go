//go:build dev

package ui

import (
	"io/fs"
	"testing/fstest"
)

// Assets returns a dummy filesystem for dev mode.
// In dev mode, the UI is proxied to the dev server, so we don't need
// to embed anything. This dummy FS satisfies the interface but won't be used.
func Assets() fs.FS {
	return fstest.MapFS{
		"200.html": &fstest.MapFile{
			Data: []byte("<!-- dev mode -->"),
		},
	}
}
