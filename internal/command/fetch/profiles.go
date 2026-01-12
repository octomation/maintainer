package fetch

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	fetchsvc "go.octolab.org/toolset/maintainer/internal/service/fetch"
)

// resolveProfiles turns the config and CLI scope into profiles with resolved
// tokens (§4.1 no-config behavior, §5 profile model and token sources).
func resolveProfiles(fs afero.Fs, fcfg *config.Fetch, opts *options, warn io.Writer) ([]fetchsvc.ResolvedProfile, error) {
	ownerFilter := toSet(opts.owners)
	profileFilter := toSet(opts.profiles)

	// No config file → single-run mode requires --owner and a token (§4.1).
	if fcfg.Source == "" {
		if len(opts.owners) == 0 {
			return nil, exit.WithUser(fmt.Errorf("not configured: provide a fetch config (fetch config init) or --owner"))
		}
		token := opts.token
		if token == "" {
			token = os.Getenv(config.DefaultTokenEnv)
		}
		if token == "" {
			return nil, exit.WithUser(fmt.Errorf("a token is required (set --token or $%s), even for public-only discovery", config.DefaultTokenEnv))
		}
		return []fetchsvc.ResolvedProfile{{Name: "default", Token: token, Owners: opts.owners}}, nil
	}

	sorted := fcfg.SortedProfiles()
	if len(sorted) == 0 {
		return nil, exit.WithUser(fmt.Errorf("config %q declares no [profiles]", fcfg.Source))
	}

	var out []fetchsvc.ResolvedProfile
	for i, np := range sorted {
		if len(profileFilter) > 0 && !profileFilter[np.Name] {
			continue // excluded by --profile scope (§5.1)
		}
		owners := np.IncludeOwners
		wildcard := config.IsAllOwners(owners)
		if len(ownerFilter) > 0 {
			if wildcard {
				owners = opts.owners // narrow "all" down to the requested owners
				wildcard = false
			} else {
				owners = intersect(owners, ownerFilter)
			}
		}
		if !wildcard && len(owners) == 0 {
			continue // nothing to do for this profile under the --owner filter
		}

		fallbackEnv := ""
		if i == 0 {
			fallbackEnv = config.DefaultTokenEnv // only the first profile defaults to GITHUB_TOKEN (§5.2)
		}
		token, warning, err := np.ResolveToken(fs, os.Getenv, fallbackEnv)
		if i == 0 && opts.token != "" {
			token, warning, err = opts.token, "", nil
		}
		if err != nil {
			return nil, exit.WithUser(fmt.Errorf("profile %q: %w", np.Name, err))
		}
		if warning != "" && warn != nil {
			fmt.Fprintf(warn, "warn: profile %q: %s\n", np.Name, warning)
		}
		out = append(out, fetchsvc.ResolvedProfile{Name: np.Name, Token: token, Owners: owners})
	}
	if len(out) == 0 {
		return nil, exit.WithUser(fmt.Errorf("no profiles selected after --profile/--owner filters"))
	}
	return out, nil
}

func toSet(values []string) map[string]bool {
	if len(values) == 0 {
		return nil
	}
	m := make(map[string]bool, len(values))
	for _, v := range values {
		m[v] = true
	}
	return m
}

func intersect(values []string, allow map[string]bool) []string {
	var out []string
	for _, v := range values {
		if allow[v] {
			out = append(out, v)
		}
	}
	return out
}
