// Package templates loads the blog's html/template set from an fs.FS and
// renders a named content template inside the shared layout.
//
// A page template is a file named <name>.html.tmpl that contains a
// `{{define "content"}}...{{end}}` block. Render(w, "<name>", data) executes
// the layout (which in turn calls `{{template "content" .}}`).
//
// The default Set is loaded from files embedded at compile time. Tests can
// build a custom Set with Load by passing any fs.FS.
package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"strings"
)

const (
	layoutFile = "layout.html.tmpl"
	pageSuffix = ".html.tmpl"
)

//go:embed *.html.tmpl
var defaultFS embed.FS

var defaultSet *Set

func init() {
	s, err := Load(defaultFS)
	if err != nil {
		panic(fmt.Errorf("templates: load default set: %w", err))
	}
	defaultSet = s
}

// Set is a parsed collection of page templates that share a common layout.
type Set struct {
	pages map[string]*template.Template
}

// Load parses the layout and every *.html.tmpl page from efs.
func Load(efs fs.FS) (*Set, error) {
	layout, err := template.ParseFS(efs, layoutFile)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", layoutFile, err)
	}
	entries, err := fs.ReadDir(efs, ".")
	if err != nil {
		return nil, fmt.Errorf("read templates dir: %w", err)
	}
	set := &Set{pages: map[string]*template.Template{}}
	for _, e := range entries {
		name := e.Name()
		if name == layoutFile || !strings.HasSuffix(name, pageSuffix) {
			continue
		}
		t, err := template.Must(layout.Clone()).ParseFS(efs, name)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		set.pages[strings.TrimSuffix(name, pageSuffix)] = t
	}
	return set, nil
}

// Render writes the page name into w, wrapped in the shared layout.
func (s *Set) Render(w io.Writer, name string, data any) error {
	t, ok := s.pages[name]
	if !ok {
		return fmt.Errorf("templates: unknown page %q", name)
	}
	return t.ExecuteTemplate(w, "layout", data)
}

// Render uses the package's default Set (loaded from embedded files).
func Render(w io.Writer, name string, data any) error {
	return defaultSet.Render(w, name, data)
}
