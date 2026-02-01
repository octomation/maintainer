package hub

import (
	"bytes"
	"strings"
	"sync"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// RenderedComment is a comment of a task with its body as HTML.
type RenderedComment struct {
	At   time.Time `json:"at"`
	URL  string    `json:"url"`
	HTML string    `json:"html"`
}

// Rendered is the HTML of a task version.
type Rendered struct {
	BodyHTML string
	Comments []RenderedComment
}

// Renderer renders the Markdown of tasks with raw HTML off and the link
// schemes limited to http, https and mailto, and caches the result by the
// task version.
type Renderer struct {
	md    goldmark.Markdown
	mu    sync.Mutex
	cache map[string]*Rendered
	order []string
	limit int
}

// NewRenderer returns a renderer whose cache keeps limit versions.
func NewRenderer(limit int) *Renderer {
	if limit <= 0 {
		limit = 512
	}
	return &Renderer{
		md: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithParserOptions(parser.WithASTTransformers(util.Prioritized(schemeFilter{}, 100))),
			goldmark.WithRendererOptions(html.WithHardWraps()),
		),
		cache: map[string]*Rendered{},
		limit: limit,
	}
}

// Render returns the HTML of the task's body and comments.
func (r *Renderer) Render(task *Task) *Rendered {
	r.mu.Lock()
	if cached, ok := r.cache[task.Version]; ok {
		r.mu.Unlock()
		return cached
	}
	r.mu.Unlock()
	out := &Rendered{BodyHTML: r.html(task.body), Comments: []RenderedComment{}}
	for _, c := range task.comments {
		out.Comments = append(out.Comments, RenderedComment{At: c.At, URL: c.URL, HTML: r.html(c.Body)})
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.cache[task.Version]; !ok {
		r.cache[task.Version] = out
		r.order = append(r.order, task.Version)
		for len(r.order) > r.limit {
			delete(r.cache, r.order[0])
			r.order = r.order[1:]
		}
	}
	return r.cache[task.Version]
}

func (r *Renderer) html(source string) string {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(source), &buf); err != nil {
		return ""
	}
	return buf.String()
}

// schemeFilter empties the destination of links, images and autolinks whose
// scheme is not http, https or mailto; relative references stay.
type schemeFilter struct{}

func (schemeFilter) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	source := reader.Source()
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Link:
			if !allowedURL(n.Destination) {
				n.Destination = nil
			}
		case *ast.Image:
			if !allowedURL(n.Destination) {
				n.Destination = nil
			}
		case *ast.AutoLink:
			if !allowedURL(n.URL(source)) {
				// An autolink has no destination of its own: keep its text.
				if parent := n.Parent(); parent != nil {
					parent.ReplaceChild(parent, n, ast.NewString(n.Label(source)))
				}
			}
		}
		return ast.WalkContinue, nil
	})
}

// AllowedURL reports whether a link destination is an http, https or mailto
// URL, or a relative reference without a scheme.
func AllowedURL(destination string) bool { return allowedURL([]byte(destination)) }

func allowedURL(destination []byte) bool {
	s := strings.TrimSpace(strings.ToLower(string(destination)))
	colon := strings.IndexByte(s, ':')
	if colon < 0 {
		return true
	}
	// A colon after a slash, a query or a fragment is not a scheme separator.
	if i := strings.IndexAny(s, "/?#"); i >= 0 && i < colon {
		return true
	}
	switch s[:colon] {
	case "http", "https", "mailto":
		return true
	}
	return false
}
