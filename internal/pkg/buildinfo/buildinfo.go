// Package buildinfo carries the build metadata that the linker sets in main
// (main.version, main.commit and main.date) to the commands that report it.
// main calls Set once, before it builds the command tree.
package buildinfo

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// Set updates the build metadata. Empty values are ignored.
func Set(version, commit, date string) {
	if version != "" {
		Version = version
	}
	if commit != "" {
		Commit = commit
	}
	if date != "" {
		Date = date
	}
}
