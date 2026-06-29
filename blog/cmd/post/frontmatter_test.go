package main

import (
	"strings"
	"testing"
)

func TestParsePost_Happy(t *testing.T) {
	src := `---
title: Hello
slug: hello
published: true
---
this is the body
across two lines
`
	fm, body, err := parsePost(src)
	if err != nil {
		t.Fatalf("parsePost: %v", err)
	}
	if fm.Title != "Hello" {
		t.Errorf("title: got %q, want %q", fm.Title, "Hello")
	}
	if fm.Slug != "hello" {
		t.Errorf("slug: got %q, want %q", fm.Slug, "hello")
	}
	if !fm.Published {
		t.Errorf("published: got false, want true")
	}
	want := "this is the body\nacross two lines\n"
	if body != want {
		t.Errorf("body: got %q, want %q", body, want)
	}
}

func TestParsePost_BlockScalarASCII(t *testing.T) {
	src := `---
title: With Art
slug: art
ascii: |
   __
  /  \
  \__/
---
body
`
	fm, _, err := parsePost(src)
	if err != nil {
		t.Fatalf("parsePost: %v", err)
	}
	want := " __\n/  \\\n\\__/"
	if fm.ASCIIArt != want {
		t.Errorf("ascii: got %q, want %q", fm.ASCIIArt, want)
	}
}

func TestParsePost_PublishedDefaultsFalse(t *testing.T) {
	src := `---
title: T
slug: s
---
b`
	fm, _, err := parsePost(src)
	if err != nil {
		t.Fatalf("parsePost: %v", err)
	}
	if fm.Published {
		t.Errorf("expected published=false by default")
	}
}

func TestParsePost_UnknownKeysIgnored(t *testing.T) {
	src := `---
title: T
slug: s
extra: whatever
also: ignored
---
body`
	if _, _, err := parsePost(src); err != nil {
		t.Errorf("unknown keys should be ignored, got error: %v", err)
	}
}

func TestParsePost_Errors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"missing opener", "title: x\nslug: y\n---\nbody", "must start with '---'"},
		{"missing closer", "---\ntitle: x\nslug: y\nbody", "missing closing '---'"},
		{"missing title", "---\nslug: y\n---\nbody", "'title:' is required"},
		{"missing slug", "---\ntitle: x\n---\nbody", "'slug:' is required"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := parsePost(c.src)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not contain %q", err.Error(), c.want)
			}
		})
	}
}
