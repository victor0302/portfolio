package main

import (
	"errors"
	"fmt"
	"strings"
)

// frontmatter is the subset of YAML-style metadata the import command
// recognizes at the top of a Markdown file:
//
//	---
//	title: Hello
//	slug: hello
//	published: true
//	ascii: |
//	  small
//	  banner
//	---
//	the body...
type frontmatter struct {
	Title     string
	Slug      string
	Published bool
	ASCIIArt  string
}

// parsePost splits src into frontmatter + body. Required fields are title
// and slug; missing them is an error. Unknown keys are ignored so files can
// carry extra metadata for other tools without breaking the importer.
func parsePost(src string) (frontmatter, string, error) {
	lines := strings.Split(src, "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != "---" {
		return frontmatter{}, "", errors.New("frontmatter: file must start with '---' on the first line")
	}

	var fm frontmatter
	i := 1
	closed := false
	for ; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if line == "---" {
			closed = true
			i++
			break
		}
		colon := strings.Index(line, ":")
		if colon == -1 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		value := strings.TrimSpace(line[colon+1:])

		if value == "|" {
			// Block scalar — consume following 2-space-indented lines.
			var block []string
			for i++; i < len(lines); i++ {
				next := strings.TrimRight(lines[i], "\r")
				if strings.HasPrefix(next, "  ") {
					block = append(block, next[2:])
				} else if next == "" {
					block = append(block, "")
				} else {
					i--
					break
				}
			}
			value = strings.TrimRight(strings.Join(block, "\n"), "\n")
		}

		switch key {
		case "title":
			fm.Title = value
		case "slug":
			fm.Slug = value
		case "published":
			fm.Published = value == "true"
		case "ascii":
			fm.ASCIIArt = value
		}
	}
	if !closed {
		return frontmatter{}, "", errors.New("frontmatter: missing closing '---' marker")
	}
	if fm.Title == "" {
		return frontmatter{}, "", errors.New("frontmatter: 'title:' is required")
	}
	if fm.Slug == "" {
		return frontmatter{}, "", errors.New("frontmatter: 'slug:' is required")
	}

	body := strings.Join(lines[i:], "\n")
	body = strings.TrimLeft(body, "\n")
	return fm, body, nil
}

func describeFrontmatter(fm frontmatter) string {
	status := "draft"
	if fm.Published {
		status = "published"
	}
	return fmt.Sprintf("slug=%s status=%s", fm.Slug, status)
}
