package markdown

import (
	"strings"
	"testing"
)

func TestToHTML(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		contains []string
	}{
		{"heading", "# Title", []string{"<h1>Title</h1>"}},
		{"paragraph + bold", "Hello **world**", []string{"<p>", "<strong>world</strong>"}},
		{"list", "- one\n- two\n", []string{"<ul>", "<li>one</li>", "<li>two</li>"}},
		{"link", "[gh](https://github.com)", []string{`<a href="https://github.com">gh</a>`}},
		{"fenced go code highlights", "```go\nfunc main() {}\n```", []string{
			`<pre`,
			`class="chroma"`,
			`<span class="kd">func</span>`, // 'func' is a keyword-declaration
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := ToHTML(c.in)
			if err != nil {
				t.Fatalf("ToHTML: %v", err)
			}
			s := string(out)
			for _, want := range c.contains {
				if !strings.Contains(s, want) {
					t.Errorf("output missing %q\ngot: %s", want, s)
				}
			}
		})
	}
}
