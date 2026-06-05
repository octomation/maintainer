package fetch

import "testing"

func TestPaint(t *testing.T) {
	if got := paint(false, "31", "x"); got != "x" {
		t.Fatalf("plain should be untouched, got %q", got)
	}
	if got := paint(true, "1;31", "error:"); got != "\x1b[1;31merror:\x1b[0m" {
		t.Fatalf("colored mismatch, got %q", got)
	}
}
