package web

import (
	"bytes"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"go.octolab.org/toolset/maintainer/internal/service/hub"
)

// Pages are the paths that serve the application document.
var Pages = []string{"/status", "/inbox", "/tablo"}

// MaxBodyBytes limits every request body.
const MaxBodyBytes = 64 << 10

// RequestedWith is the value of X-Requested-With every mutation must carry.
const RequestedWith = "maintainer"

// Options wire the handler.
type Options struct {
	// Addrs are the bound listener addresses: their loopback literals are
	// accepted as Host, with their ports.
	Addrs []net.Addr
	// Assets is the built frontend: index.html and assets/.
	Assets fs.FS
	// Index is the task index; nil serves the pages without the task API.
	Index *hub.Index
	// Checkouts caches the status rows; nil serves 503 for them.
	Checkouts *Checkouts
	// Renderer renders task bodies; nil creates one.
	Renderer *hub.Renderer
	// Label names this run: the data is served only for the host
	// <Label>.maintainer.localhost. Empty serves no data on any host.
	Label string
	// Open serves the data on the plain names too, as before the label, for
	// a machine where the user is the only account (access = "open").
	Open bool
}

type handler struct {
	Options
	mux   *http.ServeMux
	index []byte
	ips   []net.IP
	ports map[int]bool
}

// NewHandler returns the handler of the service: the pages, the assets and
// the API, behind the Host check, the mutation policy and the headers.
func NewHandler(o Options) (http.Handler, error) {
	h := &handler{Options: o, mux: http.NewServeMux(), ports: map[int]bool{}}
	if h.Renderer == nil {
		h.Renderer = hub.NewRenderer(0)
	}
	for _, addr := range o.Addrs {
		if tcp, ok := addr.(*net.TCPAddr); ok {
			h.ips = append(h.ips, tcp.IP)
			h.ports[tcp.Port] = true
		}
	}
	if o.Assets != nil {
		raw, err := fs.ReadFile(o.Assets, "index.html")
		if err != nil {
			return nil, fmt.Errorf("the frontend build has no index.html: %w", err)
		}
		if o.Open {
			// The application keeps its storage on a plain name only when
			// the page says that access is open.
			marked := bytes.Replace(raw, []byte("<html"), []byte(`<html data-access="open"`), 1)
			if bytes.Equal(marked, raw) {
				return nil, fmt.Errorf("the index.html of the frontend build has no <html> element")
			}
			raw = marked
		}
		h.index = raw
	}
	h.mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/status", http.StatusFound)
	})
	for _, page := range Pages {
		h.mux.HandleFunc("GET "+page, h.page)
	}
	h.mux.HandleFunc("GET /assets/", h.asset)
	h.mux.HandleFunc("GET /api/v0/hello", h.hello)
	h.mux.HandleFunc("GET /api/v0/checkouts", h.checkouts)
	h.mux.HandleFunc("POST /api/v0/checkouts/refresh", h.refreshCheckouts)
	h.mux.HandleFunc("GET /api/v0/tasks", h.tasks)
	h.mux.HandleFunc("GET /api/v0/tasks/{id}", h.task)
	h.mux.HandleFunc("PATCH /api/v0/tasks/{id}/triage", h.triage)
	h.mux.HandleFunc("GET /api/v0/catalog", h.catalog)
	h.mux.HandleFunc("POST /api/v0/refresh", h.refresh)
	h.mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		if allow := allowedMethods(r.URL.Path); allow != "" {
			w.Header().Set("Allow", allow)
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}
		writeError(w, http.StatusNotFound, "not_found", "no such resource")
	})
	h.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if allow := allowedMethods(r.URL.Path); allow != "" {
			w.Header().Set("Allow", allow)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	})
	return h, nil
}

// allowedMethods returns the Allow header of a known path, or "" for an
// unknown one: the catch-all handlers answer 405 instead of 404 with it.
func allowedMethods(p string) string {
	const read = "GET, HEAD"
	switch {
	case p == "/", strings.HasPrefix(p, "/assets/"):
		return read
	case p == "/api/v0/checkouts", p == "/api/v0/tasks", p == "/api/v0/catalog", p == "/api/v0/hello":
		return read
	case p == "/api/v0/refresh", p == "/api/v0/checkouts/refresh":
		return "POST"
	}
	for _, page := range Pages {
		if p == page {
			return read
		}
	}
	if rest, ok := strings.CutPrefix(p, "/api/v0/tasks/"); ok && rest != "" {
		id, tail, _ := strings.Cut(rest, "/")
		switch {
		case id == "":
			return ""
		case tail == "":
			return read
		case tail == "triage":
			return "PATCH"
		}
	}
	return ""
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; frame-ancestors 'none'")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Referrer-Policy", "no-referrer")
	access := h.hostAccess(r)
	if access == noAccess {
		writeError(w, http.StatusMisdirectedRequest, "misdirected", "the Host of the request is not this service")
		return
	}
	if h.Open {
		access = labelAccess
	}
	// Only the host of this run's label reaches the data; the plain names
	// get the pages, the assets and the hello of `maintainer service open`.
	if access != labelAccess && isAPI(r.URL.Path) && (r.Method != http.MethodGet || r.URL.Path != "/api/v0/hello") {
		writeError(w, http.StatusUnauthorized, "unauthorized", "this address has no access to the data: run `maintainer service open`")
		return
	}
	// Nothing of the hub may live in the storage of a plain name: another
	// account can serve that origin while the hub is down. Earlier releases
	// kept view state there; every page of a plain name clears it.
	if access == plainAccess && isPage(r.URL.Path) {
		header.Set("Clear-Site-Data", `"storage"`)
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
	default:
		origin := r.Header.Get("Origin")
		if origin == "" || origin == "null" || !strings.EqualFold(origin, "http://"+r.Host) {
			writeError(w, http.StatusForbidden, "forbidden", "the Origin of the request is not this service")
			return
		}
		if r.Header.Get("X-Requested-With") != RequestedWith {
			writeError(w, http.StatusForbidden, "forbidden", "X-Requested-With: maintainer is required")
			return
		}
	}
	if r.Body != nil {
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	}
	h.mux.ServeHTTP(w, r)
}

// access is what the Host of a request may reach.
type access int

const (
	// noAccess: not this service, 421.
	noAccess access = iota
	// plainAccess: maintainer.localhost, localhost or a bound loopback
	// literal: the pages, the assets and hello.
	plainAccess
	// labelAccess: <label>.maintainer.localhost of this run: everything.
	labelAccess
)

// hostAccess checks the Host of the request, with the port of the listener
// that received it: the label host of this run gets everything, the plain
// names the pages, any other host nothing. Forwarded headers are ignored.
func (h *handler) hostAccess(r *http.Request) access {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		host, port = r.Host, "80"
	}
	if local, ok := r.Context().Value(http.LocalAddrContextKey).(*net.TCPAddr); ok {
		if port != strconv.Itoa(local.Port) {
			return noAccess
		}
	} else if len(h.ports) > 0 {
		n, err := strconv.Atoi(port)
		if err != nil || !h.ports[n] {
			return noAccess
		}
	}
	name := strings.ToLower(strings.TrimSuffix(host, "."))
	if name == Host || name == "localhost" {
		return plainAccess
	}
	if label, ok := strings.CutSuffix(name, "."+Host); ok {
		if h.Label != "" && subtle.ConstantTimeCompare([]byte(label), []byte(h.Label)) == 1 {
			return labelAccess
		}
		return noAccess
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil || !ip.IsLoopback() {
		return noAccess
	}
	for _, bound := range h.ips {
		if bound.Equal(ip) {
			return plainAccess
		}
	}
	return noAccess
}

// isPage reports whether a path serves or redirects to the application
// document.
func isPage(p string) bool {
	if p == "/" {
		return true
	}
	for _, page := range Pages {
		if p == page {
			return true
		}
	}
	return false
}

// isAPI reports whether a path is under /api/, before or after cleaning.
func isAPI(p string) bool {
	clean := path.Clean("/" + p)
	return p == "/api" || strings.HasPrefix(p, "/api/") || clean == "/api" || strings.HasPrefix(clean, "/api/")
}

// Hello is the data of GET /api/v0/hello.
type Hello struct {
	// Proof is Proof(label, nonce) as hex.
	Proof string `json:"proof"`
}

// hello answers a nonce with the proof that this service knows the label of
// the run, so `maintainer service open` can tell it from another listener.
func (h *handler) hello(w http.ResponseWriter, r *http.Request) {
	nonce, err := hex.DecodeString(r.URL.Query().Get("nonce"))
	if err != nil || len(nonce) < 16 || len(nonce) > 64 {
		writeError(w, http.StatusBadRequest, "bad_request", "nonce must be 32 to 128 hex characters")
		return
	}
	if h.Label == "" {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "this service has no label")
		return
	}
	writeEnvelope(w, http.StatusOK, Envelope{Data: Hello{Proof: Proof(h.Label, nonce)}})
}

func (h *handler) page(w http.ResponseWriter, r *http.Request) {
	if h.index == nil {
		http.Error(w, "the frontend is not built", http.StatusNotFound)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "index.html", time.Time{}, strings.NewReader(string(h.index)))
}

func (h *handler) asset(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if h.Assets == nil || !strings.HasPrefix(name, "assets/") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	file, err := h.Assets.Open(name)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	// The bundled typeface: not every system's MIME table knows .woff2, and
	// the answer is nosniff.
	if path.Ext(name) == ".woff2" {
		w.Header().Set("Content-Type", "font/woff2")
	}
	http.ServeContent(w, r, info.Name(), time.Time{}, seeker)
}

// Envelope wraps every API answer.
type Envelope struct {
	Generation  uint64     `json:"generation"`
	CollectedAt *time.Time `json:"collected_at"`
	Refreshing  bool       `json:"refreshing"`
	Warnings    []string   `json:"warnings"`
	Data        any        `json:"data"`
}

// ErrorBody is the JSON of an API error.
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail names an API error.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorBody{Error: ErrorDetail{Code: code, Message: message}})
}

func writeEnvelope(w http.ResponseWriter, status int, envelope Envelope) {
	if envelope.Warnings == nil {
		envelope.Warnings = []string{}
	}
	writeJSON(w, status, envelope)
}
