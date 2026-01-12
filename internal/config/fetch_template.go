package config

// FetchConfigTemplate is the documented TOML template written by
// `maintainer fetch config init` (§4.2). It is intentionally a 1:1 mirror of
// the schema documented in the plan, with comments preserved.
const FetchConfigTemplate = `# maintainer fetch configuration
# Location: ./fetch.toml or $XDG_CONFIG_HOME/maintainer/fetch.toml
# Format is picked by extension; a fetch.yaml form is a 1:1 translation.

[defaults]
root         = "."        # checkout root; defaults to the current working directory
path         = "{{.Visibility}}/{{.Owner}}/{{.Repo}}"
clone_url    = "ssh"      # "ssh" | "https"
concurrency  = 4
# state_file = "/path/to/state.json"  # default: $XDG_STATE_HOME/maintainer/fetch/state.json

[filters]
# Filters gate only NEW clone/adopt decisions; tracked repos stay tracked.
exclude_archived  = false
exclude_forks     = false
exclude_templates = false

[profiles.primary]
token_env      = "GITHUB_TOKEN"
# include_owners is an allowlist AND the set of owners to query.
# Use ["*"] (or omit it) to mean "you + every org you are a member of".
include_owners = ["acme-user", "acme", "acme-labs", "acme-tools"]

[profiles.bot]
token_env      = "ACME_BOT_TOKEN"
include_owners = ["acme-bot"]
clone_url      = "https"   # per-profile override; HTTPS+PAT for private repos

# [[owners]]
# name = "acme"
# path = "{{.Visibility}}/{{.Owner}}/{{.Repo}}"   # explicit per-owner override

# [[repos]]
# match = { owner = "acme-user", name = "dotfiles" }
# path  = "~/.dotfiles"

# [[repos]]
# match = { id = 12345678 }                       # id-match survives rename
# path  = "~/Code/special"

# [[repos]]
# match  = { id = 99999999 }                      # silence a confirmed orphan
# ignore = true
`
