package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/mattn/go-isatty"
)

// Output formats for the plan (§3.1 --format).
const (
	FormatHuman = "human"
	FormatJSON  = "json"
)

// Reporter owns stdout (plan + summary) and stderr (logs + progress); it
// switches between human and JSON per --format (§9). Leveled verbosity uses
// log/slog, the deliberate polish-milestone addition (§2.1, §11.1).
type Reporter struct {
	out       io.Writer
	err       io.Writer
	format    string
	verbosity int
	quiet     bool
	color     bool // stderr is an interactive terminal
	colorOut  bool // stdout is an interactive terminal
	debug     *slog.Logger
}

// NewReporter builds a Reporter. An unknown format falls back to human.
// verbosity maps to §11.1 levels (0=warn, 1=info, 2=debug, 3=trace); quiet
// suppresses everything below error.
func NewReporter(out, err io.Writer, format string, verbosity int, quiet bool) *Reporter {
	if format != FormatJSON {
		format = FormatHuman
	}
	level := slog.LevelWarn
	switch {
	case quiet:
		level = slog.LevelError
	case verbosity >= 2:
		level = slog.LevelDebug
	case verbosity == 1:
		level = slog.LevelInfo
	}
	debug := slog.New(slog.NewTextHandler(err, &slog.HandlerOptions{Level: level}))
	return &Reporter{
		out: out, err: err, format: format, verbosity: verbosity, quiet: quiet,
		color: wantColor(err), colorOut: wantColor(out), debug: debug,
	}
}

// wantColor reports whether ANSI colour is appropriate for w: only when w is an
// interactive terminal and NO_COLOR is unset. A piped/redirected stream or a
// test buffer gets plain output.
func wantColor(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if f, ok := w.(*os.File); ok {
		return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}
	return false
}

// paint wraps s in an ANSI SGR code when on; otherwise returns s untouched.
func paint(on bool, code, s string) string {
	if !on {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

// Format returns the active output format.
func (r *Reporter) Format() string { return r.format }

// Render writes the plan to stdout. applied=true switches the trailing hint
// from "run with --apply" to the executed summary.
func (r *Reporter) Render(p Plan, applied bool) error {
	if r.format == FormatJSON {
		return r.renderJSON(p)
	}
	return r.renderHuman(p, applied)
}

// human plan symbols (§7.3).
var symbols = map[Kind]string{
	KindClone:        "+",
	KindFetch:        "~",
	KindMove:         "↻",
	KindRelocate:     "↪",
	KindUpdateRemote: "±",
	KindAdopt:        "=",
	KindOrphan:       "!",
	KindConflict:     "✗",
	KindNoop:         " ",
}

// drift reports whether an action is shown as an individual human line. Routine
// fetches and plain noops are collapsed into the summary (§7.3).
func drift(a Action) bool {
	switch a.Kind {
	case KindFetch:
		return false
	case KindNoop:
		return a.Flag != "" // flagged noops (inaccessible/legal-hold) stay visible
	default:
		return true
	}
}

func (r *Reporter) renderHuman(p Plan, applied bool) error {
	var b strings.Builder

	for _, d := range p.Discoveries {
		fmt.Fprintf(&b, "profile=%-9s discovered=%d repos%s\n", d.Profile, d.Count, endpoints(d))
	}
	if p.StatePath != "" {
		fmt.Fprintf(&b, "state             loaded %d records from %s\n", p.StateCount, r.short(p, p.StatePath))
	}
	b.WriteByte('\n')

	fmt.Fprintf(&b, "plan: %d repos total\n", len(p.Actions))
	for _, a := range p.Actions {
		if !drift(a) {
			continue
		}
		r.writeAction(&b, p, a)
	}

	s := p.Summary()
	b.WriteByte('\n')
	fmt.Fprintf(&b, "summary: %s\n", summaryLine(s))
	switch {
	case applied && s.Errors > 0:
		fmt.Fprintf(&b, "%s\n", paint(r.colorOut, "1;31", fmt.Sprintf("applied with %d error(s)", s.Errors)))
	case applied:
		b.WriteString(paint(r.colorOut, "32", "applied") + "\n")
	case actionable(s):
		b.WriteString(paint(r.colorOut, "36", "run with --apply to execute") + "\n")
	default:
		b.WriteString(paint(r.colorOut, "32", "no drift") + "\n")
	}

	_, err := io.WriteString(r.out, b.String())
	return err
}

func (r *Reporter) writeAction(b *strings.Builder, p Plan, a Action) {
	sym := symbols[a.Kind]
	name := a.Owner + "/" + a.Name
	switch a.Kind {
	case KindMove:
		fmt.Fprintf(b, "  %s move       %-28s → renamed on GitHub: %s/%s\n", sym, a.FromName, a.Owner, a.Name)
		fmt.Fprintf(b, "               from %s\n", r.short(p, a.FromPath))
		fmt.Fprintf(b, "               to   %s\n", r.short(p, a.ToPath))
		if a.UpdateRemote {
			b.WriteString("               + update remote.origin.url\n")
		}
	case KindRelocate:
		fmt.Fprintf(b, "  %s relocate   %-28s → moved on disk\n", sym, name)
		fmt.Fprintf(b, "               from %s\n", r.short(p, a.FromPath))
		fmt.Fprintf(b, "               to   %s\n", r.short(p, a.ToPath))
	case KindUpdateRemote:
		fmt.Fprintf(b, "  %s update     %-28s → %s\n", sym, name, a.RemoteURL)
	case KindOrphan:
		fmt.Fprintf(b, "  %s orphan     %-28s → %s\n", sym, name, a.Reason)
		fmt.Fprintf(b, "               at %s\n", r.short(p, a.Path))
	case KindConflict:
		fmt.Fprintf(b, "  %s conflict   %-28s → %s\n", sym, name, a.Reason)
	case KindNoop:
		fmt.Fprintf(b, "  %s %-10s %-28s → %s\n", sym, string(a.Kind), name, a.Reason)
	default:
		fmt.Fprintf(b, "  %s %-10s %-28s → %s\n", sym, string(a.Kind), name, r.short(p, a.Path))
	}
}

func (r *Reporter) renderJSON(p Plan) error {
	doc := jsonPlan{
		PlanID:      p.ID,
		GeneratedAt: p.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z"),
		Discoveries: p.Discoveries,
		Summary:     p.Summary(),
	}
	for _, a := range p.Actions {
		doc.Actions = append(doc.Actions, toJSONAction(a))
	}
	enc := json.NewEncoder(r.out)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

// Logf writes a lifecycle marker to stderr; shown at the default level and
// suppressed only by --quiet (§11.1).
func (r *Reporter) Logf(format string, args ...any) {
	if r.quiet {
		return
	}
	fmt.Fprintf(r.err, format+"\n", args...)
}

// Infof writes a line shown at -v and above (§11.1 info level).
func (r *Reporter) Infof(format string, args ...any) {
	if r.quiet || r.verbosity < 1 {
		return
	}
	fmt.Fprintf(r.err, format+"\n", args...)
}

// Debugf writes a structured breadcrumb shown at -vv and above (§11.1).
func (r *Reporter) Debugf(msg string, args ...any) {
	r.debug.Debug(msg, args...)
}

// Errorf writes an error line; always shown, even under --quiet (§3.1). On an
// interactive terminal it is painted red so it stands out in the stream.
func (r *Reporter) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(r.err, "%s %s\n", paint(r.color, "1;31", "error:"), paint(r.color, "31", msg))
}

func (r *Reporter) short(p Plan, path string) string {
	if p.Root != "" && strings.HasPrefix(path, p.Root) {
		return "<root>" + strings.TrimPrefix(path, p.Root)
	}
	return path
}

func endpoints(d DiscoverySummary) string {
	if len(d.Endpoints) == 0 {
		return ""
	}
	parts := make([]string, 0, len(d.Endpoints))
	for _, e := range d.Endpoints {
		parts = append(parts, e.Endpoint)
	}
	sort.Strings(parts)
	return " (" + strings.Join(parts, ", ") + ")"
}

func summaryLine(s Summary) string {
	return fmt.Sprintf(
		"clone=%d fetch=%d move=%d relocate=%d update_remote=%d adopt=%d orphan=%d noop=%d conflict=%d errors=%d",
		s.Clone, s.Fetch, s.Move, s.Relocate, s.UpdateRemote, s.Adopt, s.Orphan, s.Noop, s.Conflict, s.Errors,
	)
}

func actionable(s Summary) bool {
	return s.Clone+s.Move+s.Relocate+s.UpdateRemote+s.Adopt+s.Fetch > 0
}

type jsonPlan struct {
	PlanID      string             `json:"plan_id"`
	GeneratedAt string             `json:"generated_at"`
	Discoveries []DiscoverySummary `json:"discoveries"`
	Actions     []jsonAction       `json:"actions"`
	Summary     Summary            `json:"summary"`
}

type jsonAction struct {
	Kind     string `json:"kind"`
	ID       int64  `json:"id"`
	Owner    string `json:"owner,omitempty"`
	Name     string `json:"name,omitempty"`
	Path     string `json:"path,omitempty"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	FromPath string `json:"from_path,omitempty"`
	ToPath   string `json:"to_path,omitempty"`
	Flag     string `json:"flag,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

func toJSONAction(a Action) jsonAction {
	out := jsonAction{
		Kind: string(a.Kind), ID: a.ID, Owner: a.Owner, Name: a.Name,
		Path: a.Path, Flag: a.Flag, Reason: a.Reason,
	}
	if a.Kind == KindMove {
		out.From = a.FromName
		out.To = a.Owner + "/" + a.Name
		out.FromPath = a.FromPath
		out.ToPath = a.ToPath
		out.Path = ""
	}
	if a.Kind == KindRelocate {
		out.FromPath = a.FromPath
		out.ToPath = a.ToPath
		out.Path = ""
	}
	return out
}
