package fetch

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.octolab.org/toolset/maintainer/internal/config"
)

func TestParseRemote(t *testing.T) {
	type remote struct{ host, owner, name, transport string }
	tests := []struct {
		raw  string
		want remote
		ok   bool
	}{
		{"git@github.com:octomation/maintainer.git", remote{"github.com", "octomation", "maintainer", config.TransportSSH}, true},
		{"github.com:octomation/maintainer", remote{"github.com", "octomation", "maintainer", config.TransportSSH}, true},
		{"git@github.com:/octomation/maintainer.git", remote{"github.com", "octomation", "maintainer", config.TransportSSH}, true},
		{"ssh://git@github.com/octomation/maintainer.git", remote{"github.com", "octomation", "maintainer", config.TransportSSH}, true},
		{"ssh://git@github.com:2222/octomation/maintainer.git", remote{"github.com:2222", "octomation", "maintainer", config.TransportSSH}, true},
		{"git+ssh://github.com/octomation/maintainer", remote{"github.com", "octomation", "maintainer", config.TransportSSH}, true},
		{"git://github.com/octomation/maintainer.git", remote{"github.com", "octomation", "maintainer", config.TransportSSH}, true},
		{"https://github.com/octomation/maintainer.git", remote{"github.com", "octomation", "maintainer", config.TransportHTTPS}, true},
		{"https://token@github.com/octomation/maintainer", remote{"github.com", "octomation", "maintainer", config.TransportHTTPS}, true},
		{"HTTP://github.com/octomation/maintainer", remote{"github.com", "octomation", "maintainer", config.TransportHTTPS}, true},
		{"git@[::1]:octomation/maintainer.git", remote{"[::1]", "octomation", "maintainer", config.TransportSSH}, true},
		{"first.last@github-work:octomation/maintainer.git", remote{"github-work", "octomation", "maintainer", config.TransportSSH}, true},
		// A local path is SSH as well: the transport is HTTPS only for http(s).
		{"octomation/maintainer.git", remote{"", "octomation", "maintainer", config.TransportSSH}, true},
		{"file:///octomation/maintainer", remote{"", "octomation", "maintainer", config.TransportSSH}, true},

		{"", remote{}, false},
		{"https://github.com/octomation", remote{}, false},
		{"https://github.com/octomation/maintainer/tree", remote{}, false},
		{"https://github.com/octomation/", remote{}, false},
		{"git@github.com:maintainer.git", remote{}, false},
		{"/home/user/octomation/maintainer", remote{}, false},
		{`C:\src\octomation\maintainer`, remote{}, false},
		{"svn://github.com/octomation/maintainer", remote{}, false},
		{"ssh://git@github.com:octomation/maintainer.git", remote{}, false},
	}
	for _, tc := range tests {
		t.Run(tc.raw, func(t *testing.T) {
			host, owner, name, transport, ok := parseRemote(tc.raw)
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.want, remote{host, owner, name, transport})
		})
	}
}

func TestSameRepository(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"git@github.com:octomation/maintainer.git", "https://github.com/octomation/maintainer", true},
		{"git@GitHub.com:Octomation/Maintainer.git", "https://github.com/octomation/maintainer.git", true},
		{"git@github.com:octomation/maintainer.git", "git@github.com:octomation/indexit.git", false},
		{"git@github.com:octomation/maintainer.git", "git@gitlab.com:octomation/maintainer.git", false},
		{"git@github.com:octomation/maintainer.git", "", false},
		{"", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.a+" "+tc.b, func(t *testing.T) {
			assert.Equal(t, tc.want, sameRepository(tc.a, tc.b))
			assert.Equal(t, tc.want, sameRepository(tc.b, tc.a))
		})
	}
}
