package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.octolab.org/toolset/maintainer/internal/service/hub"
	"go.octolab.org/toolset/maintainer/internal/service/hub/mirror"
	"go.octolab.org/toolset/maintainer/internal/service/status"
)

// CheckoutRow is a status row with its display cells.
type CheckoutRow struct {
	status.Row
	Cells       []string `json:"cells"`
	DisplayPath string   `json:"display_path"`
	Placeholder bool     `json:"placeholder"`
}

// CheckoutsData is the data of GET /api/v0/checkouts.
type CheckoutsData struct {
	Rows    []CheckoutRow  `json:"rows"`
	Summary status.Summary `json:"summary"`
	Total   int            `json:"total"`
	Sort    []SortSpec     `json:"sort"`
	Query   string         `json:"query"`
	Error   string         `json:"error,omitempty"`
}

// SortSpec is one key of the sort chain.
type SortSpec struct {
	Column string `json:"column"`
	Desc   bool   `json:"desc"`
}

// TaskDetail is the data of GET /api/v0/tasks/{id}.
type TaskDetail struct {
	*hub.Task
	BodyHTML   string                `json:"body_html"`
	Comments   []hub.RenderedComment `json:"comments"`
	Provenance Provenance            `json:"provenance"`
}

// Provenance says where a task was read from.
type Provenance struct {
	Path string `json:"path"`
	Root string `json:"root"`
}

// TriagePatch is the body of PATCH /api/v0/tasks/{id}/triage: an omitted
// field is unchanged, null removes the label, a value replaces it.
type TriagePatch struct {
	Impact *string `json:"impact"`
	Effort *string `json:"effort"`
}

func parseSort(raw string) ([]status.SortKey, []SortSpec, error) {
	var keys []status.SortKey
	specs := []SortSpec{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		desc := strings.HasPrefix(part, "-")
		name := strings.TrimPrefix(part, "-")
		column, ok := status.ParseColumn(name)
		if !ok {
			return nil, nil, errors.New("unknown sort column " + strconv.Quote(name))
		}
		keys = append(keys, status.SortKey{Column: column, Desc: desc})
		specs = append(specs, SortSpec{Column: strings.ToLower(name), Desc: desc})
	}
	return keys, specs, nil
}

func (h *handler) checkouts(w http.ResponseWriter, r *http.Request) {
	if h.Checkouts == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "the checkouts are not collected by this service")
		return
	}
	keys, specs, err := parseSort(r.URL.Query().Get("sort"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	h.Checkouts.Ensure()
	rows, collectedAt, ok, refreshing, failure := h.Checkouts.Snapshot()
	if !ok {
		envelope := Envelope{Refreshing: refreshing}
		if failure != "" {
			envelope.Warnings = []string{failure}
		}
		writeEnvelope(w, http.StatusAccepted, envelope)
		return
	}
	query := r.URL.Query().Get("q")
	visible := status.Query(rows, keys, query)
	shown := make([]status.Row, 0, len(visible))
	data := CheckoutsData{Rows: make([]CheckoutRow, 0, len(visible)), Sort: specs, Query: query, Error: failure}
	for _, index := range visible {
		row := rows[index]
		shown = append(shown, row)
		data.Rows = append(data.Rows, CheckoutRow{Row: row, Cells: status.Cells(row), DisplayPath: row.DisplayPath(), Placeholder: row.IsPlaceholder()})
	}
	data.Summary = status.Summarize(shown)
	data.Total = len(rows)
	at := collectedAt
	envelope := Envelope{CollectedAt: &at, Refreshing: refreshing, Data: data}
	if failure != "" {
		envelope.Warnings = []string{failure}
	}
	writeEnvelope(w, http.StatusOK, envelope)
}

func (h *handler) refreshCheckouts(w http.ResponseWriter, r *http.Request) {
	if h.Checkouts == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "the checkouts are not collected by this service")
		return
	}
	h.Checkouts.Refresh()
	_, collectedAt, ok, _, _ := h.Checkouts.Snapshot()
	envelope := Envelope{Refreshing: true}
	if ok {
		at := collectedAt
		envelope.CollectedAt = &at
	}
	writeEnvelope(w, http.StatusAccepted, envelope)
}

// snapshot answers 202 and returns nil before the first index.
func (h *handler) snapshot(w http.ResponseWriter) *hub.Snapshot {
	if h.Index == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "the tasks are not indexed by this service")
		return nil
	}
	snap := h.Index.Snapshot()
	if snap == nil {
		writeEnvelope(w, http.StatusAccepted, Envelope{Refreshing: true})
		return nil
	}
	return snap
}

func (h *handler) envelope(snap *hub.Snapshot, data any) Envelope {
	at := snap.IndexedAt
	return Envelope{Generation: snap.Generation, CollectedAt: &at, Refreshing: h.Index.Refreshing(), Warnings: snap.Warnings, Data: data}
}

func parseQuery(values url.Values) (hub.Query, error) {
	q := hub.Query{
		Owners: values["owner"], Repos: values["repo"], Milestones: values["milestone"], Labels: values["label"],
		Types: values["type"], Assignees: values["assignee"], Columns: values["column"],
		State: values.Get("state"), Q: values.Get("q"), Sort: values.Get("sort"),
	}
	for name, target := range map[string]*int{"limit": &q.Limit, "offset": &q.Offset} {
		if raw := values.Get(name); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil {
				return q, errors.New(name + " must be a number")
			}
			*target = n
		}
	}
	if err := q.Validate(); err != nil {
		return q, err
	}
	return q, nil
}

func (h *handler) tasks(w http.ResponseWriter, r *http.Request) {
	q, err := parseQuery(r.URL.Query())
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	snap := h.snapshot(w)
	if snap == nil {
		return
	}
	writeEnvelope(w, http.StatusOK, h.envelope(snap, snap.Query(q)))
}

func (h *handler) lookup(w http.ResponseWriter, r *http.Request) (*hub.Snapshot, *hub.Task) {
	key, err := hub.DecodeID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return nil, nil
	}
	snap := h.snapshot(w)
	if snap == nil {
		return nil, nil
	}
	task := snap.Task(key)
	if task == nil {
		writeError(w, http.StatusNotFound, "not_found", "no such task")
		return nil, nil
	}
	return snap, task
}

func (h *handler) detail(task *hub.Task) TaskDetail {
	rendered := h.Renderer.Render(task)
	return TaskDetail{Task: task, BodyHTML: rendered.BodyHTML, Comments: rendered.Comments,
		Provenance: Provenance{Path: task.Path(), Root: task.Root().Path}}
}

func (h *handler) task(w http.ResponseWriter, r *http.Request) {
	snap, task := h.lookup(w, r)
	if task == nil {
		return
	}
	w.Header().Set("ETag", strconv.Quote(task.Version))
	writeEnvelope(w, http.StatusOK, h.envelope(snap, h.detail(task)))
}

func (h *handler) triage(w http.ResponseWriter, r *http.Request) {
	if media, _, _ := strings.Cut(r.Header.Get("Content-Type"), ";"); strings.TrimSpace(strings.ToLower(media)) != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "the body must be application/json")
		return
	}
	match := r.Header.Get("If-Match")
	if match == "" {
		writeError(w, http.StatusPreconditionRequired, "precondition_required", "If-Match with the ETag of the task is required")
		return
	}
	version, err := strconv.Unquote(match)
	if err != nil || strings.HasPrefix(match, "W/") {
		writeError(w, http.StatusBadRequest, "bad_request", "If-Match must be the quoted strong ETag")
		return
	}
	triage, err := decodeTriage(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if h.Index == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "the tasks are not indexed by this service")
		return
	}
	key, err := hub.DecodeID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	task, snap, err := h.Index.Patch(key, version, triage)
	switch {
	case err == nil:
	case errors.Is(err, mirror.ErrDirectorySync):
		// The file is replaced and the index follows it; the sync failed.
		writeError(w, http.StatusInternalServerError, "internal", "the label was written and the file was replaced, but the directory could not be synced: "+err.Error())
		return
	case errors.Is(err, hub.ErrNotIndexed):
		writeError(w, http.StatusServiceUnavailable, "unavailable", "the index is not ready")
		return
	case errors.Is(err, hub.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "no such task")
		return
	case errors.Is(err, hub.ErrDuplicateMirror):
		writeError(w, http.StatusConflict, "conflict", "the task has duplicate mirrors; edit the file of one of them")
		return
	case errors.Is(err, mirror.ErrVersionMismatch):
		writeError(w, http.StatusPreconditionFailed, "precondition_failed", "the file changed; reload the task")
		return
	case errors.Is(err, mirror.ErrNotEditable):
		writeError(w, http.StatusUnprocessableEntity, "unprocessable", err.Error())
		return
	case errors.Is(err, mirror.ErrNotRegular):
		writeError(w, http.StatusUnprocessableEntity, "unprocessable", err.Error())
		return
	default:
		writeError(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	w.Header().Set("ETag", strconv.Quote(task.Version))
	writeEnvelope(w, http.StatusOK, h.envelope(snap, h.detail(task)))
}

// decodeTriage reads the patch: unknown and duplicate keys are errors, and
// so are values outside the label vocabularies.
func decodeTriage(r *http.Request) (mirror.Triage, error) {
	dec := json.NewDecoder(r.Body)
	malformed := errors.New("malformed JSON body")
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return mirror.Triage{}, malformed
	}
	var triage mirror.Triage
	seen := map[string]bool{}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return mirror.Triage{}, malformed
		}
		key, _ := tok.(string)
		var change *mirror.Change
		var allowed []string
		switch key {
		case "impact":
			change, allowed = &triage.Impact, []string{"high", "medium", "low"}
		case "effort":
			change, allowed = &triage.Effort, []string{"easy", "medium", "hard"}
		default:
			return mirror.Triage{}, errors.New("unknown key " + strconv.Quote(key))
		}
		if seen[key] {
			return mirror.Triage{}, errors.New("duplicate key " + strconv.Quote(key))
		}
		seen[key] = true
		var value *string
		if err := dec.Decode(&value); err != nil {
			return mirror.Triage{}, errors.New(key + " must be a string or null")
		}
		if value == nil {
			*change = mirror.Change{Set: true}
			continue
		}
		found := false
		for _, v := range allowed {
			found = found || v == *value
		}
		if !found {
			return mirror.Triage{}, errors.New(key + " must be one of " + strings.Join(allowed, ", ") + " or null")
		}
		*change = mirror.Change{Set: true, Value: *value}
	}
	if tok, err := dec.Token(); err != nil || tok != json.Delim('}') {
		return mirror.Triage{}, malformed
	}
	if _, err := dec.Token(); err != io.EOF {
		return mirror.Triage{}, malformed
	}
	return triage, nil
}

func (h *handler) catalog(w http.ResponseWriter, r *http.Request) {
	snap := h.snapshot(w)
	if snap == nil {
		return
	}
	writeEnvelope(w, http.StatusOK, h.envelope(snap, snap.Catalog))
}

func (h *handler) refresh(w http.ResponseWriter, r *http.Request) {
	if h.Index == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "the tasks are not indexed by this service")
		return
	}
	h.Index.Refresh()
	envelope := Envelope{Refreshing: true}
	if snap := h.Index.Snapshot(); snap != nil {
		at := snap.IndexedAt
		envelope.Generation, envelope.CollectedAt, envelope.Warnings = snap.Generation, &at, snap.Warnings
	}
	writeEnvelope(w, http.StatusAccepted, envelope)
}
