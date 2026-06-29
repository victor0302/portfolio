// Package markdown converts post bodies (stored as Markdown) into HTML.
//
// Fenced code blocks are syntax-highlighted by goldmark-highlighting (chroma
// under the hood) with class-based output. Colors live in blog.css so the
// light and dark themes can style tokens independently.
package markdown

import (
	"bytes"
	"fmt"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		highlighting.NewHighlighting(
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
			),
		),
	),
)

// ToHTML renders src as HTML.
func ToHTML(src string) ([]byte, error) {
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return nil, fmt.Errorf("convert markdown: %w", err)
	}
	return buf.Bytes(), nil
}
