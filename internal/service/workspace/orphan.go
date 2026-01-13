package workspace

// Orphan reasons describe management, independently of Git tracking status.
const (
	OrphanDuplicatePin = "duplicate-pin"
	OrphanOutOfScope   = "out-of-scope"
	OrphanRemoteGone   = "remote-gone"
)

func OrphanDescription(reason, active string) string {
	switch reason {
	case OrphanDuplicatePin:
		return "inactive checkout; active pin: " + active + "; local copy retained"
	case OrphanOutOfScope:
		return "known checkout outside current management scope; local copy retained"
	case OrphanRemoteGone:
		return "repository not found by GitHub ID (404); local clone retained"
	default:
		return reason
	}
}
