package header

import (
	"net/http"

	"go.octolab.org/toolkit/protocol/http/header"
)

func Set(header http.Header) Setter {
	return Setter(header)
}

type Setter http.Header

func (s Setter) NoCache() Setter {
	h := http.Header(s)
	h.Set(header.CacheControl, "no-cache")
	h.Add(header.CacheControl, "no-store")
	h.Add(header.CacheControl, "must-revalidate")
	h.Set("Pragma", "no-cache")
	h.Set("Expires", "0")
	return s
}
