// Package exit carries process exit codes through the error chain so that
// command handlers can express the CLI contract (§3.3 of the fetch plan)
// without importing os, and the entrypoint maps a typed error to a code.
package exit

// Process exit codes used across the maintainer CLI.
const (
	OK        = 0 // clean finish (including "no drift")
	Transport = 1 // transport / Git / state error (also the default)
	User      = 2 // user input error (bad config, bad flag combo, missing token)
	Partial   = 3 // apply requested but at least one per-repo action failed
)

// Coder is implemented by errors that carry a process exit code.
type Coder interface {
	ExitCode() int
}

type coded struct {
	code int
	err  error
}

func (e coded) Error() string { return e.err.Error() }
func (e coded) Unwrap() error { return e.err }
func (e coded) ExitCode() int { return e.code }

// With wraps err so the entrypoint resolves it to the given exit code.
// A nil err yields nil, so callers can wrap unconditionally.
func With(code int, err error) error {
	if err == nil {
		return nil
	}
	return coded{code: code, err: err}
}

// WithUser wraps err as a user input error (exit code 2).
func WithUser(err error) error { return With(User, err) }

// WithPartial wraps err as a partial-apply failure (exit code 3).
func WithPartial(err error) error { return With(Partial, err) }
