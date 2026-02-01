package mirror

import (
	"regexp"
	"strings"
	"time"
)

// Comment is one trailing comment block of an issue file:
// <!-- <timestamp> <url>\n…\n-->.
type Comment struct {
	At   time.Time
	URL  string
	Body string
}

var commentOpen = regexp.MustCompile(`^<!--\s+(\S+)\s+(\S+)\s*$`)

// SplitComments separates the trailing comment blocks from the body. A block
// starts with a line `<!-- <timestamp> <url>` outside fenced code and ends
// with a line `-->`; only the blocks after the last other content are
// comments, every other HTML comment stays in the body.
func SplitComments(body string) (string, []Comment) {
	lines := strings.Split(body, "\n")
	type block struct {
		start, end int // line indexes, inclusive
		comment    Comment
	}
	var blocks []block
	var fence string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimRight(line, "\r")
		if fence != "" {
			if closesFence(trimmed, fence) {
				fence = ""
			}
			continue
		}
		if marker := fenceMarker(trimmed); marker != "" {
			fence = marker
			continue
		}
		m := commentOpen.FindStringSubmatch(trimmed)
		if m == nil {
			continue
		}
		at := ParseTime(m[1])
		if at == nil {
			continue
		}
		end := -1
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimRight(lines[j], "\r") == "-->" {
				end = j
				break
			}
		}
		if end < 0 {
			continue
		}
		text := make([]string, 0, end-i-1)
		for _, l := range lines[i+1 : end] {
			text = append(text, strings.TrimRight(l, "\r"))
		}
		blocks = append(blocks, block{start: i, end: end, comment: Comment{At: *at, URL: m[2], Body: strings.Join(text, "\n")}})
		i = end
	}
	// Keep the blocks at the tail: from the last one backwards, with only
	// blank lines between them and after the last.
	cut := len(lines)
	var comments []Comment
	for k := len(blocks) - 1; k >= 0; k-- {
		b := blocks[k]
		if !blank(lines[b.end+1 : cut]) {
			break
		}
		comments = append([]Comment{b.comment}, comments...)
		cut = b.start
	}
	if cut == len(lines) {
		return body, nil
	}
	return strings.TrimRight(strings.Join(lines[:cut], "\n"), " \t\r\n") + "\n", comments
}

func blank(lines []string) bool {
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			return false
		}
	}
	return true
}

// closesFence reports whether a line closes the fence opened with marker:
// the same character, at least as many of them, at most three spaces of
// indentation and nothing but whitespace after.
func closesFence(line, marker string) bool {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > 3 || marker == "" {
		return false
	}
	n := 0
	for n < len(trimmed) && trimmed[n] == marker[0] {
		n++
	}
	return n >= len(marker) && strings.TrimSpace(trimmed[n:]) == ""
}

// fenceMarker returns the opening fence of a fenced code block, or "".
func fenceMarker(line string) string {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > 3 {
		return ""
	}
	for _, ch := range []string{"```", "~~~"} {
		if strings.HasPrefix(trimmed, ch) {
			n := 0
			for n < len(trimmed) && trimmed[n] == ch[0] {
				n++
			}
			return trimmed[:n]
		}
	}
	return ""
}
