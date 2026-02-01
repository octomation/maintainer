package giturl

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	user := url.User
	tests := []struct {
		in   string
		want url.URL
		str  string // the URL reserialized; empty means in, "-" skips the check
	}{
		// Ported from github.com/whilp/git-urls, after git-clone(1).
		{"user@host.xz:path/to/repo.git/", url.URL{Scheme: "ssh", User: user("user"), Host: "host.xz", Path: "path/to/repo.git/"}, "ssh://user@host.xz/path/to/repo.git/"},
		{"host.xz:path/to/repo.git/", url.URL{Scheme: "ssh", Host: "host.xz", Path: "path/to/repo.git/"}, "ssh://host.xz/path/to/repo.git/"},
		{"host.xz:/path/to/repo.git/", url.URL{Scheme: "ssh", Host: "host.xz", Path: "/path/to/repo.git/"}, "ssh://host.xz/path/to/repo.git/"},
		{"host.xz:path/to/repo-with_specials.git/", url.URL{Scheme: "ssh", Host: "host.xz", Path: "path/to/repo-with_specials.git/"}, "ssh://host.xz/path/to/repo-with_specials.git/"},
		{"git://host.xz/path/to/repo.git/", url.URL{Scheme: "git", Host: "host.xz", Path: "/path/to/repo.git/"}, ""},
		{"git://host.xz:1234/path/to/repo.git/", url.URL{Scheme: "git", Host: "host.xz:1234", Path: "/path/to/repo.git/"}, ""},
		{"http://host.xz/path/to/repo.git/", url.URL{Scheme: "http", Host: "host.xz", Path: "/path/to/repo.git/"}, ""},
		{"http://host.xz:1234/path/to/repo.git/", url.URL{Scheme: "http", Host: "host.xz:1234", Path: "/path/to/repo.git/"}, ""},
		{"https://host.xz/path/to/repo.git/", url.URL{Scheme: "https", Host: "host.xz", Path: "/path/to/repo.git/"}, ""},
		{"https://host.xz:1234/path/to/repo.git/", url.URL{Scheme: "https", Host: "host.xz:1234", Path: "/path/to/repo.git/"}, ""},
		{"ftp://host.xz/path/to/repo.git/", url.URL{Scheme: "ftp", Host: "host.xz", Path: "/path/to/repo.git/"}, ""},
		{"ftp://host.xz:1234/path/to/repo.git/", url.URL{Scheme: "ftp", Host: "host.xz:1234", Path: "/path/to/repo.git/"}, ""},
		{"ftps://host.xz/path/to/repo.git/", url.URL{Scheme: "ftps", Host: "host.xz", Path: "/path/to/repo.git/"}, ""},
		{"ftps://host.xz:1234/path/to/repo.git/", url.URL{Scheme: "ftps", Host: "host.xz:1234", Path: "/path/to/repo.git/"}, ""},
		{"rsync://host.xz/path/to/repo.git/", url.URL{Scheme: "rsync", Host: "host.xz", Path: "/path/to/repo.git/"}, ""},
		{"ssh://user@host.xz:1234/path/to/repo.git/", url.URL{Scheme: "ssh", User: user("user"), Host: "host.xz:1234", Path: "/path/to/repo.git/"}, ""},
		{"ssh://host.xz:1234/path/to/repo.git/", url.URL{Scheme: "ssh", Host: "host.xz:1234", Path: "/path/to/repo.git/"}, ""},
		{"ssh://host.xz/path/to/repo.git/", url.URL{Scheme: "ssh", Host: "host.xz", Path: "/path/to/repo.git/"}, ""},
		{"git+ssh://host.xz/path/to/repo.git/", url.URL{Scheme: "git+ssh", Host: "host.xz", Path: "/path/to/repo.git/"}, ""},
		{"/path/to/repo.git/", url.URL{Scheme: "file", Path: "/path/to/repo.git/"}, "file:///path/to/repo.git/"},
		{"file:///path/to/repo.git/", url.URL{Scheme: "file", Path: "/path/to/repo.git/"}, ""},
		{"https://host.xz/organization/repo.git?ref=", url.URL{Scheme: "https", Host: "host.xz", Path: "/organization/repo.git", RawQuery: "ref="}, ""},
		{"https://host.xz/organization/repo.git?ref=test", url.URL{Scheme: "https", Host: "host.xz", Path: "/organization/repo.git", RawQuery: "ref=test"}, ""},
		{"https://host.xz/organization/repo.git?ref=feature/test", url.URL{Scheme: "https", Host: "host.xz", Path: "/organization/repo.git", RawQuery: "ref=feature/test"}, ""},
		{"git@host.xz:organization/repo.git?ref=test", url.URL{Scheme: "ssh", User: user("git"), Host: "host.xz", Path: "organization/repo.git", RawQuery: "ref=test"}, "ssh://git@host.xz/organization/repo.git?ref=test"},
		{"git@host.xz:organization/repo.git?ref=feature/test", url.URL{Scheme: "ssh", User: user("git"), Host: "host.xz", Path: "organization/repo.git", RawQuery: "ref=feature/test"}, "ssh://git@host.xz/organization/repo.git?ref=feature/test"},
		{"https://user:password@host.xz/organization/repo.git/", url.URL{Scheme: "https", User: url.UserPassword("user", "password"), Host: "host.xz", Path: "/organization/repo.git/"}, ""},
		{"https://user:password@host.xz/organization/repo.git?ref=test", url.URL{Scheme: "https", User: url.UserPassword("user", "password"), Host: "host.xz", Path: "/organization/repo.git", RawQuery: "ref=test"}, ""},
		{"https://user:password@host.xz/organization/repo.git?ref=feature/test", url.URL{Scheme: "https", User: url.UserPassword("user", "password"), Host: "host.xz", Path: "/organization/repo.git", RawQuery: "ref=feature/test"}, ""},

		// Remotes as GitHub prints them.
		{"git@github.com:octomation/maintainer.git", url.URL{Scheme: "ssh", User: user("git"), Host: "github.com", Path: "octomation/maintainer.git"}, "ssh://git@github.com/octomation/maintainer.git"},
		{"https://github.com/octomation/maintainer.git", url.URL{Scheme: "https", Host: "github.com", Path: "/octomation/maintainer.git"}, ""},
		{"HTTPS://github.com/octomation/maintainer", url.URL{Scheme: "https", Host: "github.com", Path: "/octomation/maintainer"}, "https://github.com/octomation/maintainer"},

		// A user with dots and hyphens, an alias host, a port-like path.
		{"first.last-name@host.xz:a/b", url.URL{Scheme: "ssh", User: user("first.last-name"), Host: "host.xz", Path: "a/b"}, "ssh://first.last-name@host.xz/a/b"},
		{"github-work:a/b.git", url.URL{Scheme: "ssh", Host: "github-work", Path: "a/b.git"}, "ssh://github-work/a/b.git"},
		{"host.xz:22/a/b", url.URL{Scheme: "ssh", Host: "host.xz", Path: "22/a/b"}, "ssh://host.xz/22/a/b"},
		{"host.xz::a/b", url.URL{Scheme: "ssh", Host: "host.xz", Path: ":a/b"}, "ssh://host.xz/:a/b"},
		{"a@b@host.xz:x/y", url.URL{Scheme: "ssh", User: user("a@b"), Host: "host.xz", Path: "x/y"}, "ssh://a%40b@host.xz/x/y"},
		{"git@host.xz:a/b+c~d.git", url.URL{Scheme: "ssh", User: user("git"), Host: "host.xz", Path: "a/b+c~d.git"}, "ssh://git@host.xz/a/b+c~d.git"},
		// Without "://" a scheme is a host, as Git reads it.
		{"https:host.xz/a/b", url.URL{Scheme: "ssh", Host: "https", Path: "host.xz/a/b"}, "ssh://https/host.xz/a/b"},
		{"host.xz:", url.URL{Scheme: "ssh", Host: "host.xz"}, "ssh://host.xz"},

		// IPv6 hosts.
		{"git@[::1]:a/b.git", url.URL{Scheme: "ssh", User: user("git"), Host: "[::1]", Path: "a/b.git"}, "ssh://git@[::1]/a/b.git"},
		{"[fe80::1]:a/b", url.URL{Scheme: "ssh", Host: "[fe80::1]", Path: "a/b"}, "ssh://[fe80::1]/a/b"},
		{"ssh://git@[::1]:2222/a/b.git", url.URL{Scheme: "ssh", User: user("git"), Host: "[::1]:2222", Path: "/a/b.git"}, ""},
		{"https://[2001:db8::1]/a/b", url.URL{Scheme: "https", Host: "[2001:db8::1]", Path: "/a/b"}, ""},

		// Local paths.
		{"", url.URL{Scheme: "file"}, "file:"},
		{"repo", url.URL{Scheme: "file", Path: "repo"}, "-"},
		{"../a/b.git", url.URL{Scheme: "file", Path: "../a/b.git"}, "-"},
		{"./a:b", url.URL{Scheme: "file", Path: "./a:b"}, "-"},
		{"/tmp/a://b", url.URL{Scheme: "file", Path: "/tmp/a://b"}, "file:///tmp/a://b"},
		{"file://host.xz/a/b", url.URL{Scheme: "file", Host: "host.xz", Path: "/a/b"}, ""},
		{`C:\src\repo`, url.URL{Scheme: "file", Path: `C:\src\repo`}, "-"},
		{"C:/src/repo", url.URL{Scheme: "file", Path: "C:/src/repo"}, "-"},
		{"c:", url.URL{Scheme: "file", Path: "c:"}, "-"},
		{":a/b", url.URL{Scheme: "file", Path: ":a/b"}, "-"},
		{"git@:a/b", url.URL{Scheme: "file", Path: "git@:a/b"}, "-"},
		{"[::1/a:b", url.URL{Scheme: "file", Path: "[::1/a:b"}, "-"},
		{"[::1]a:b", url.URL{Scheme: "file", Path: "[::1]a:b"}, "-"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := Parse(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.want, *got)
			switch tc.str {
			case "-":
			case "":
				assert.Equal(t, tc.in, got.String())
			default:
				assert.Equal(t, tc.str, got.String())
			}
		})
	}
}

func TestParseRejects(t *testing.T) {
	for _, in := range []string{
		"svn://host.xz/a/b",                   // unknown transport
		"ssh+git://host.xz/a/b",               // not a scheme the parser knows
		"ssh://git@github.com:owner/repo.git", // scp-like path after a scheme
		"https://host.xz:port/a/b",            // invalid port
		"https://host.xz/a%zz",                // invalid escape
	} {
		t.Run(in, func(t *testing.T) {
			_, err := Parse(in)
			assert.Error(t, err)
		})
	}
}
