// Package giturl parses the remote URLs Git accepts: URLs with a transport
// scheme, the scp-like syntax and local paths.
package giturl

import (
	"fmt"
	"net/url"
	"strings"
)

// transports are the schemes accepted in a URL with "://".
var transports = map[string]bool{
	"ssh": true, "git+ssh": true, "git": true, "http": true, "https": true,
	"ftp": true, "ftps": true, "rsync": true, "file": true,
}

// Parse follows Git's rules. A string with "://" after a scheme is a URL and
// must name a known transport. Otherwise a colon before the first slash makes
// it scp-like, [user@]host:path, with the "ssh" scheme; a query after "?" is
// kept apart from the path. Anything else, a Windows drive path included, is
// a local path with the "file" scheme.
func Parse(raw string) (*url.URL, error) {
	if scheme, _, ok := strings.Cut(raw, "://"); ok && isScheme(scheme) {
		u, err := url.Parse(raw)
		if err != nil {
			return nil, err
		}
		if !transports[u.Scheme] {
			return nil, fmt.Errorf("scheme %q is not a valid transport", u.Scheme)
		}
		return u, nil
	}
	if u, ok := parseScp(raw); ok {
		return u, nil
	}
	return &url.URL{Scheme: "file", Path: raw}, nil
}

// parseScp splits [user@]host:path, where the host may be a bracketed IPv6
// address. As in Git, the syntax needs a colon before the first slash.
func parseScp(raw string) (*url.URL, bool) {
	colon := strings.IndexByte(raw, ':')
	slash := strings.IndexByte(raw, '/')
	if colon < 0 || slash >= 0 && slash < colon || isDrive(raw, colon) {
		return nil, false
	}
	u := &url.URL{Scheme: "ssh"}
	rest := raw
	if at := strings.LastIndexByte(raw[:colon], '@'); at >= 0 {
		u.User, rest = url.User(raw[:at]), raw[at+1:]
	}
	var path string
	if strings.HasPrefix(rest, "[") {
		host, after, ok := strings.Cut(rest, "]")
		if !ok || !strings.HasPrefix(after, ":") {
			return nil, false
		}
		u.Host, path = host+"]", after[1:]
	} else {
		u.Host, path, _ = strings.Cut(rest, ":")
	}
	if u.Host == "" || strings.Contains(u.Host, "/") {
		return nil, false
	}
	u.Path, u.RawQuery, _ = strings.Cut(path, "?")
	return u, true
}

// isDrive reports a Windows drive prefix such as C:\ or C:/.
func isDrive(raw string, colon int) bool {
	return colon == 1 && isLetter(raw[0]) && (len(raw) == 2 || raw[2] == '\\' || raw[2] == '/')
}

// isScheme reports RFC 3986 scheme syntax: a letter, then letters, digits,
// "+", "-" or ".".
func isScheme(s string) bool {
	if s == "" || !isLetter(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !isLetter(c) && (c < '0' || c > '9') && c != '+' && c != '-' && c != '.' {
			return false
		}
	}
	return true
}

func isLetter(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}
