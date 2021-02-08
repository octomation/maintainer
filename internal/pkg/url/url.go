package url

import (
	"net/http"
	"net/url"
	"path"
)

func MustParse(raw string) *URL {
	u, err := Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}

func Parse(raw string) (*URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &URL{origin: *u, query: u.Query()}, nil
}

type URL struct {
	origin url.URL
	query  url.Values
}

func (u URL) AddPath(p string) URL {
	u.origin.Path = path.Join(u.origin.Path, p)
	return u
}

func (u URL) SetPath(p string) URL {
	u.origin.Path = p
	return u
}

func (u URL) AddQueryParam(k, v string) URL {
	u.query = url.Values(http.Header(u.query).Clone())
	u.query.Add(k, v)
	return u
}

func (u URL) SetQueryParam(k, v string) URL {
	u.query = url.Values(http.Header(u.query).Clone())
	u.query.Set(k, v)
	return u
}

func (u URL) DelQueryParam(k string) URL {
	u.query = url.Values(http.Header(u.query).Clone())
	u.query.Del(k)
	return u
}

func (u URL) String() string {
	u.origin.RawQuery = u.query.Encode()
	return u.origin.String()
}
