// Schema for .github/settings.json: the repository settings that cannot be
// derived at runtime. Validate with `make config-vet`.
//
// Derived, never stored here: owner/name (git remote, GITHUB_REPOSITORY),
// the default branch (origin/HEAD, the event payload, the API), the site URL
// (actions/configure-pages, checked against pages), the Homebrew tap
// (.goreleaser.yml).
//
// The top level stays open for GitHub REST shaped sections, e.g. repository,
// pages or rulesets, once they are managed from this file.
package settings

#Settings: {
	// Release policy, read by .github/scripts/release.mjs from the tagged
	// commit: the pre-push hook, `make release-check` and cd.yml.
	"x-release"?: #Release

	// Secrets the workflows need, by name. Values never live here.
	// `make doctor` checks that each one exists and is passed by a workflow,
	// and prints how to create a missing one.
	secrets?: [Name=string & =~"^[A-Z_][A-Z0-9_]*$"]: #Secret

	// GitHub Pages as the REST API reports it (GET /repos/{owner}/{repo}/pages).
	// The docs build bakes the site URL in, so cd.docs.yml and cd.yml refuse
	// to build when Pages disagrees, and `make doctor` reports the drift.
	pages?: #Pages

	...
}

#Release: {
	// Which tags are releases. A tag matching refs/tags/v* (the cd.yml
	// trigger) but not this pattern is refused.
	tag_pattern: string | *"^v\\d+\\.\\d+\\.\\d+(-[0-9A-Za-z.-]+)?$"

	// The curated note of a release; {tag} is replaced with the tag name.
	// It is both a docs page and the GitHub release body.
	notes: string | *"docs/content/changelog/{tag}.md"

	// Where a release tag must point. The first rule whose regexp matches
	// the tag wins; $1… expand its groups. With no match, the default branch.
	// Example for maintenance lines: {match: "^v(\\d+)\\.", branch: "v$1"}.
	branches: [...{
		match:  string
		branch: string
	}] | *[]
}

#Pages: {
	// The custom domain; absent means https://<owner>.github.io/<repo>/.
	// Settings → Pages → Custom domain, not a CNAME file: a workflow
	// deployment ignores it.
	cname?: string & =~"^[a-z0-9.-]+$"

	https_enforced?: bool

	// The docs are built by cd.docs.yml, not from a branch.
	build_type?: "workflow"
}

#Secret: {
	// Where the secret is expected to be set.
	scope: "repository" | "organization" | "environment"
}

#Settings
