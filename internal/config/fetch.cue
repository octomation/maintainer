// Schema for the fetch config, fetch.toml or fetch.yaml, read by
// `maintainer fetch` and `maintainer status`. It is the contract: the loader
// rejects every key and table outside it. `make config-vet` checks the
// fixtures in testdata against it, the template of `maintainer fetch config
// init` and the guide examples among them, and the Go tests check that the
// loader accepts and rejects the same fixtures.
//
// Validate a config with `cue vet -c fetch.cue fetch.toml`.
//
// The loader accepts what the schema accepts and checks more: the layout must
// stay within the root, templates may use only the documented fields, pins
// are literal paths, [workspace] and the legacy defaults.root/path do not mix,
// and a repos match names an id or both owner and name.
package config

#Fetch: {
	// The checkout tree shared with status. Do not mix it with the legacy
	// defaults.root and defaults.path; without either, the defaults below
	// apply.
	workspace?: #Workspace

	defaults?: #Defaults

	// Filters skip new clones and adoptions, of pinned checkouts too;
	// repositories already tracked stay tracked.
	filters?: #Filters

	// Profiles by name, each a token with the owners it discovers. The first
	// one in alphabetical order reads $GITHUB_TOKEN when it names no token.
	profiles?: [Name=string]: #Profile

	// Per-owner overrides of the layout and the transport.
	owners?: [...#Owner]

	// Per-repository rules, matched by id or by owner and name.
	repos?: [...#Repo]
}

#Workspace: {
	// The checkout root; a relative root is resolved from the current
	// working directory.
	root: string | *"."

	// The layout of managed checkouts, relative to root.
	path: string | *"{{.Visibility}}/{{.Owner}}/{{.Repo}}"

	// Existing checkouts or trees of checkouts kept in place, as literal
	// paths: relative to root, absolute or ~/.
	pins: [...string] | *[]
}

#Defaults: {
	// The legacy spelling of workspace.root and workspace.path.
	root?: string
	path?: string

	clone_url: #Transport | *"ssh"

	// 0 means the default.
	concurrency: int & >=0 | *4

	// Default: $XDG_STATE_HOME/maintainer/fetch/state.json, falling back to
	// ~/.local/state/maintainer/fetch/state.json.
	state_file?: string
}

#Filters: {
	exclude_archived:  bool | *false
	exclude_forks:     bool | *false
	exclude_templates: bool | *false
}

#Profile: {
	// Token sources in the order they are tried. token_file must not be
	// readable by others (0600 or stricter); an inline token is a warning.
	token_file?: string
	token_env?:  string
	token?:      string

	// The owners to discover and the allowlist after discovery, compared
	// case-insensitively. Omitted or "*" means you and every organization you
	// are a member of.
	include_owners?: [...string]

	// Overrides defaults.clone_url for this profile.
	clone_url?: #Transport
}

#Owner: {
	name: string & !=""

	// Overrides workspace.path for this owner, relative to root.
	path?:      string
	clone_url?: #Transport
}

#Repo: {
	// An id survives renames and transfers; owner and name match together.
	match: #RepoMatch

	// Pins an existing checkout: absolute, ~/ or relative to root.
	path?:      string
	clone_url?: #Transport

	// Silences the whole pipeline for this repository.
	ignore: bool | *false
}

#RepoMatch: {
	// 0 is unset: owner and name match instead.
	id?:    int
	owner?: string
	name?:  string
}

// An empty transport is unset: the next level decides.
#Transport: "ssh" | "https" | ""

#Fetch
