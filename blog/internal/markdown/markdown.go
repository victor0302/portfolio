// Package markdown converts post bodies (stored as Markdown) into HTML.
package markdown

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
)

var md = goldmark.New()

// ToHTML renders src as HTML using goldmark's default extensions.
func ToHTML(src string) ([]byte, error) {
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return nil, fmt.Errorf("convert markdown: %w", err)
	}
	return buf.Bytes(), nil
}
