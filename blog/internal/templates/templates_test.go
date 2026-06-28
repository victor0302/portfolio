package templates

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadAndRender(t *testing.T) {
	efs := fstest.MapFS{
		"layout.html.tmpl": &fstest.MapFile{Data: []byte(
			`{{define "layout"}}<html><body><nav>Home</nav>{{template "content" .}}</body></html>{{end}}`,
		)},
		"hello.html.tmpl": &fstest.MapFile{Data: []byte(
			`{{define "content"}}<p>Hi {{.Name}}</p>{{end}}`,
		)},
	}

	set, err := Load(efs)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	var buf bytes.Buffer
	if err := set.Render(&buf, "hello", struct{ Name string }{Name: "Vic"}); err != nil {
		t.Fatalf("Render: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "<nav>Home</nav>") {
		t.Errorf("expected layout nav in output, got: %s", out)
	}
	if !strings.Contains(out, "<p>Hi Vic</p>") {
		t.Errorf("expected content block in output, got: %s", out)
	}
}

func TestRender_UnknownPage(t *testing.T) {
	efs := fstest.MapFS{
		"layout.html.tmpl": &fstest.MapFile{Data: []byte(
			`{{define "layout"}}{{template "content" .}}{{end}}`,
		)},
	}
	set, err := Load(efs)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := set.Render(&bytes.Buffer{}, "nope", nil); err == nil {
		t.Fatalf("expected error for unknown page, got nil")
	}
}

func TestLoad_MissingLayoutFails(t *testing.T) {
	efs := fstest.MapFS{
		"hello.html.tmpl": &fstest.MapFile{Data: []byte(
			`{{define "content"}}hi{{end}}`,
		)},
	}
	if _, err := Load(efs); err == nil {
		t.Fatalf("expected error when layout is missing, got nil")
	}
}

func TestDefaultSet_RendersEmbeddedLayout(t *testing.T) {
	// The embedded layout should at minimum parse; rendering a page name that
	// doesn't yet exist returns "unknown page" — that's expected until ticket
	// 22/23 land. This test guards against the init() panicking on a malformed
	// embedded layout.
	if defaultSet == nil {
		t.Fatal("defaultSet not initialized by init()")
	}
	err := Render(&bytes.Buffer{}, "definitely-not-a-page", nil)
	if err == nil || !strings.Contains(err.Error(), "unknown page") {
		t.Fatalf("expected 'unknown page' error, got %v", err)
	}
}
