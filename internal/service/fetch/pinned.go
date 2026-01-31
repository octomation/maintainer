package fetch

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"text/template"
	"text/template/parse"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// ErrUnresolvedPin reports a per-repo path that references a field that
// neither the configuration, the match rule nor a state record supplies.
var ErrUnresolvedPin = errors.New("unresolved pin")

// PinnedPath resolves an explicit repo path, including an old-name rule during
// a rename. The applicable rule wins when it renders from sufficient context;
// the persisted pin is the fallback, which keeps name-based rules safe on
// subsequent runs.
func (pr *PathResolver) PinnedPath(snap github.RepoSnapshot, rec *state.Record) (string, error) {
	rule, ok := pr.cnf.RepoOverride(snap.ID, snap.Owner, snap.Name)
	if !ok && rec != nil {
		rule, ok = pr.cnf.RepoOverride(rec.ID, rec.OwnerLogin, rec.Name)
	}
	if ok && rule.Path != "" {
		path, err := pr.renderPin(rule, snap, rec)
		if errors.Is(err, ErrUnresolvedPin) && rec != nil && rec.PinSource != "workspace" && rec.PinnedPath != "" {
			return rec.PinnedPath, nil
		}
		return path, err
	}
	if rec != nil {
		if rec.PinSource != "workspace" && rec.PinnedPath != "" {
			return rec.PinnedPath, nil
		}
		if pr.scope != nil && pr.scope.Pinned(rec.Path) {
			return rec.Path, nil
		}
	}
	return "", nil
}

// renderPin renders a per-repo path only from sufficient context: the
// snapshot completed from the record and the match rule. A snapshot built
// from {id, owner, name} alone never stands in for unknown metadata.
func (pr *PathResolver) renderPin(rule config.Repo, snap github.RepoSnapshot, rec *state.Record) (string, error) {
	fields, err := templateFields(rule.Path)
	if err != nil {
		return "", err
	}
	snap, known := pinContext(snap, rec)
	var missing []string
	for field, ok := range known {
		if !ok && (fields[field] || fields["."]) {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		slices.Sort(missing)
		return "", fmt.Errorf("%w: path %q references unknown %s", ErrUnresolvedPin, rule.Path, strings.Join(missing, ", "))
	}
	return pr.renderer.Render(rule.Path, snap, true)
}

// pinContext completes snap with the identity and the metadata from the
// record, unless GitHub supplied them; a name-based rule matches only a
// snapshot or a record that carries its owner and name. It reports which
// template fields are known.
func pinContext(snap github.RepoSnapshot, rec *state.Record) (github.RepoSnapshot, map[string]bool) {
	if rec != nil {
		if snap.Owner == "" && snap.Name == "" {
			snap.Owner, snap.Name = rec.OwnerLogin, rec.Name
		}
		if snap.Visibility == "" {
			snap.Visibility, snap.DefaultBranch = github.Visibility(rec.Visibility), rec.DefaultBranch
			snap.IsFork, snap.IsTemplate, snap.IsArchived = rec.IsFork, rec.IsTemplate, rec.ArchivedOnGitHub
		}
	}
	metadata := snap.Visibility != ""
	return snap, map[string]bool{
		"Owner":         snap.Owner != "",
		"Repo":          snap.Name != "",
		"Visibility":    metadata,
		"DefaultBranch": metadata && snap.DefaultBranch != "",
		"IsFork":        metadata,
		"IsTemplate":    metadata,
		"IsArchived":    metadata,
	}
}

// templateFields lists the fields a path template references; "." stands for
// the whole context, as a bare dot or $ at the top level references it. The
// dot of a with block, and the dot and $ of a called template, are the value
// passed to it.
func templateFields(tmpl string) (map[string]bool, error) {
	t, err := template.New("path").Funcs(tmplFuncs).Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("parse path template %q: %w", tmpl, err)
	}
	fields := make(map[string]bool)
	// A holder is "" for the whole context, a field name, or "?" for a value
	// computed from fields that are already listed.
	type scope struct{ dot, root string }
	use := func(holder string) {
		switch holder {
		case "":
			fields["."] = true
		case "?":
		default:
			fields[holder] = true
		}
	}
	field := func(holder, name string) {
		if holder == "" {
			fields[name] = true
		} else {
			use(holder)
		}
	}
	value := func(pipe *parse.PipeNode, sc scope) string {
		if pipe == nil || len(pipe.Cmds) != 1 || len(pipe.Cmds[0].Args) != 1 {
			return "?"
		}
		holder, name := "?", ""
		switch arg := pipe.Cmds[0].Args[0].(type) {
		case *parse.DotNode:
			return sc.dot
		case *parse.FieldNode:
			holder, name = sc.dot, arg.Ident[0]
			if len(arg.Ident) > 1 {
				return "?"
			}
		case *parse.VariableNode:
			if arg.Ident[0] != "$" || len(arg.Ident) > 2 {
				return "?"
			}
			if len(arg.Ident) == 1 {
				return sc.root
			}
			holder, name = sc.root, arg.Ident[1]
		}
		if holder == "" {
			return name
		}
		return "?"
	}
	called := make(map[[2]string]bool)
	var walk func(node parse.Node, sc scope)
	walk = func(node parse.Node, sc scope) {
		switch n := node.(type) {
		case *parse.ListNode:
			if n != nil {
				for _, child := range n.Nodes {
					walk(child, sc)
				}
			}
		case *parse.ActionNode:
			walk(n.Pipe, sc)
		case *parse.PipeNode:
			if n != nil {
				for _, cmd := range n.Cmds {
					walk(cmd, sc)
				}
			}
		case *parse.CommandNode:
			for _, arg := range n.Args {
				walk(arg, sc)
			}
		case *parse.ChainNode:
			walk(n.Node, sc)
		case *parse.FieldNode:
			field(sc.dot, n.Ident[0])
		case *parse.VariableNode:
			if n.Ident[0] == "$" && len(n.Ident) > 1 {
				field(sc.root, n.Ident[1])
			} else if n.Ident[0] == "$" {
				use(sc.root)
			}
		case *parse.DotNode:
			use(sc.dot)
		case *parse.IfNode:
			walk(n.Pipe, sc)
			walk(n.List, sc)
			walk(n.ElseList, sc)
		case *parse.RangeNode:
			walk(n.Pipe, sc)
			walk(n.List, scope{"?", sc.root})
			walk(n.ElseList, sc)
		case *parse.WithNode:
			walk(n.Pipe, sc)
			walk(n.List, scope{value(n.Pipe, sc), sc.root})
			walk(n.ElseList, sc)
		case *parse.TemplateNode:
			// Passing dot or a field is no use of it; the called template decides.
			inner := value(n.Pipe, sc)
			if inner == "?" {
				walk(n.Pipe, sc)
			}
			if tmpl := t.Lookup(n.Name); tmpl != nil && tmpl.Tree != nil && !called[[2]string{n.Name, inner}] {
				called[[2]string{n.Name, inner}] = true
				walk(tmpl.Root, scope{inner, inner})
			}
		}
	}
	walk(t.Root, scope{})
	return fields, nil
}

// Authorised refuses to turn remembered paths into additional scan roots.
// A rule that cannot be rendered still marks the repository as explicitly
// managed, so that the plan reports it instead of an orphan.
func (pr *PathResolver) Authorised(rec state.Record, snap github.RepoSnapshot) bool {
	if snap.ID == 0 {
		snap = github.RepoSnapshot{ID: rec.ID, Owner: rec.OwnerLogin, Name: rec.Name}
	}
	if pin, err := pr.PinnedPath(snap, &rec); err != nil || pin != "" {
		return true
	}
	if rec.PinSource == "workspace" {
		return false
	} // removing a pin never enables moves
	return pr.scope != nil && pr.scope.Contains(rec.Path)
}

func (pr *PathResolver) WorkspacePin(snap github.RepoSnapshot, rec *state.Record) bool {
	rule, ok := pr.cnf.RepoOverride(snap.ID, snap.Owner, snap.Name)
	if !ok && rec != nil {
		rule, ok = pr.cnf.RepoOverride(rec.ID, rec.OwnerLogin, rec.Name)
	}
	if ok && rule.Path != "" {
		return false
	}
	return rec != nil && (rec.PinSource == "workspace" || rec.PinnedPath == "") && pr.scope.Pinned(rec.Path)
}

// An explicit path selects the active clone even when an old duplicate exists.
// It never moves or creates directories: changing the pin selects an existing,
// identity-verified checkout, or reports a conflict.
func planPinned(base Action, target string, rec *state.Record, clones []DiskClone) *Action {
	base.Path = target
	c, found := cloneAt(clones, target)
	if !found || c.Origins > 1 {
		base.Kind = KindConflict
		base.Reason = "pinned path is missing or is not an unambiguous clone of this repository"
		return &base
	}
	if c.Transport != "" {
		base.Transport = c.Transport
	}
	base.RemoteURL = canonicalURL(base.Transport, base.Owner, base.Name)
	base.Record = recordFromSnapshot(*base.Snapshot, target, base.RemoteURL, base.Transport)
	base.Record.PinnedPath = target
	if rec != nil {
		base.Record.FirstSeen = rec.FirstSeen
		base.Record.PreviousPaths = append([]string(nil), rec.PreviousPaths...)
		if rec.Path != "" && rec.Path != target {
			found := false
			for _, path := range base.Record.PreviousPaths {
				found = found || path == rec.Path
			}
			if !found {
				base.Record.PreviousPaths = append(base.Record.PreviousPaths, rec.Path)
			}
		}
	}
	drift := c.RemoteURL != "" && c.RemoteURL != base.RemoteURL
	switch {
	case rec == nil || rec.Path != target:
		// Adoption only adds: it records the checkout with its observed
		// origin, and a drifted origin is the next apply's update_remote.
		base.Kind = KindAdopt
		base.staleOrigin = c.rebound
		if c.RemoteURL != "" {
			base.RemoteURL = c.RemoteURL
			base.Record.RemoteURL = c.RemoteURL
		}
	case drift:
		base.Kind = KindUpdateRemote
		base.UpdateRemote = true
	default:
		base.Kind = KindFetch
	}
	return &base
}
