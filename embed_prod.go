//go:build embed

package main

import (
	"embed"
	"io/fs"
)

//go:embed web/dist
var embeddedWebFS embed.FS

func webFS() fs.FS {
	sub, err := fs.Sub(embeddedWebFS, "web/dist")
	if err != nil {
		panic("embed: web/dist sub: " + err.Error())
	}
	return sub
}
