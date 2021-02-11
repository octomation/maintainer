package vanity

// MetaImport represents the parsed
// <meta name="go-import" content="prefix vcs reporoot" />
// tags from HTML files.
//
// See cmd/go/internal/get/vcs.go (metaImport).
type MetaImport struct {
	Prefix, VCS, RepoRoot string
}

type MetaSource struct {
	URL, Dir, File string
}

type Meta struct {
	Package string
	Import  MetaImport
	Source  MetaSource
}
