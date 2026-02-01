package web

import (
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
		h.index = raw
	}
	h.mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/status", http.StatusFound)
	})
	for _, page := range Pages {
		h.mux.HandleFunc("GET "+page, h.page)
	}
	h.mux.HandleFunc("GET /assets/", h.asset)
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
	case p == "/api/v0/checkouts", p == "/api/v0/tasks", p == "/api/v0/catalog":
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
	if !h.hostAllowed(r) {
		writeError(w, http.StatusMisdirectedRequest, "misdirected", "the Host of the request is not this service")
		return
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

// hostAllowed accepts maintainer.localhost, localhost and the bound loopback
// literals, with the port of the listener that received the request.
// Forwarded headers are ignored.
func (h *handler) hostAllowed(r *http.Request) bool {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		host, port = r.Host, "80"
	}
	if local, ok := r.Context().Value(http.LocalAddrContextKey).(*net.TCPAddr); ok {
		if port != strconv.Itoa(local.Port) {
			return false
		}
	} else if len(h.ports) > 0 {
		n, err := strconv.Atoi(port)
		if err != nil || !h.ports[n] {
			return false
		}
	}
	name := strings.ToLower(strings.TrimSuffix(host, "."))
	if name == Host || name == "localhost" {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil || !ip.IsLoopback() {
		return false
	}
	for _, bound := range h.ips {
		if bound.Equal(ip) {
			return true
		}
	}
	return false
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
