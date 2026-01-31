package fetch

import (
	"encoding/json"
	"errors"
	"fmt"
	iofs "io/fs"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// stateCommand implements `maintainer fetch state {show,prune}` (§3).
func stateCommand(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "state",
		Short: "inspect and tidy the local state file",
		Args:  cobra.NoArgs,
	}

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "dump the current state file as JSON",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, st, release, err := openState(cmd, opts)
			if err != nil {
				return err
			}
			defer func() { _ = release() }()
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(st)
		},
	}

	pruneCmd := &cobra.Command{
		Use:   "prune",
		Short: "drop records whose disk path is missing",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, st, release, err := openState(cmd, opts)
			if err != nil {
				return err
			}
			defer func() { _ = release() }()
			fs := afero.NewOsFs()
			var dropped, unchecked int
			kept := st.Repos[:0]
			for _, rec := range st.Repos {
				// Only a path known to be absent is gone; any other failure,
				// such as an unreadable parent, leaves the record as it is.
				_, err := fs.Stat(rec.Path)
				switch {
				case err == nil:
					kept = append(kept, rec)
				case errors.Is(err, iofs.ErrNotExist):
					dropped++
					fmt.Fprintf(cmd.ErrOrStderr(), "prune: id=%d %s/%s (missing %s)\n", rec.ID, rec.OwnerLogin, rec.Name, rec.Path)
				default:
					unchecked++
					kept = append(kept, rec)
					var pathErr *iofs.PathError
					if errors.As(err, &pathErr) {
						err = pathErr.Err
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "prune: id=%d %s/%s kept, cannot check %s: %v\n", rec.ID, rec.OwnerLogin, rec.Name, rec.Path, err)
				}
			}
			st.Repos = kept
			if err := store.Save(st); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "pruned %d record(s)\n", dropped)
			if unchecked > 0 {
				return fmt.Errorf("kept %d record(s) whose path could not be checked", unchecked)
			}
			return nil
		},
	}

	command.AddCommand(showCmd, pruneCmd)
	return command
}

// openState resolves the state store, acquires the advisory lock, and loads
// the state. The caller must defer the returned release func. It selects the
// config as fetch does, so a discovered fetch.toml's state_file is respected
// and --config="" means the default state location.
func openState(cmd *cobra.Command, opts *options) (*state.Store, *state.State, func() error, error) {
	fs := afero.NewOsFs()
	fcfg, err := config.LoadFetch(fs, configPath(cmd, fs, opts))
	if err != nil {
		return nil, nil, nil, exit.WithUser(err)
	}
	store, err := newStore(fs, fcfg)
	if err != nil {
		return nil, nil, nil, err
	}
	release, err := store.Lock()
	if err != nil {
		return nil, nil, nil, exit.WithUser(err)
	}
	st, err := store.Load()
	if err != nil {
		_ = release()
		return nil, nil, nil, err
	}
	return store, st, release, nil
}
