// Package static serves the blog's CSS (and any future static assets)
// from files embedded at compile time.
package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed *.css *.js
var files embed.FS

// FS returns the embedded static asset tree.
func FS() fs.FS {
	return files
}

// Handler serves the embedded files. Mount it under /static/ with
// http.StripPrefix so requests like /static/blog.css resolve to blog.css.
func Handler() http.Handler {
	return http.FileServer(http.FS(files))
}
