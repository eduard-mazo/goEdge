//go:build !embed

package main

import "io/fs"

// webFS returns nil in dev mode; Vite dev server handles static files.
func webFS() fs.FS { return nil }
