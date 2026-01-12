package contribution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	stdio "io"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"
	"gopkg.in/yaml.v2"

	"go.octolab.org/toolset/maintainer/internal/pkg/io"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

var packer io.Packer

// this is internals, so, we know what we are doing
func init() {
	packer = io.NewPacker()
	packer.Register(
		func(w io.Writer) io.Encoder { return json.NewEncoder(w) },
		func(r io.Reader) io.Decoder { return json.NewDecoder(r) },
		".json",
	)
	packer.Register(
		func(w io.Writer) io.Encoder { return yaml.NewEncoder(w) },
		func(r io.Reader) io.Decoder { return yaml.NewDecoder(r) },
		".yml", ".yaml",
	)
}

func NewFileSource(src afero.Fs, path string) *fileSource {
	return &fileSource{
		Provider: src,
		Path:     path,
	}
}

type fileSource struct {
	Provider afero.Fs
	Path     string

	data HeatMap
}

func (src fileSource) Location() string {
	return fmt.Sprintf("file:%s", src.Path)
}

func (src *fileSource) Fetch(_ context.Context) (HeatMap, error) {
	if src.data != nil {
		return src.data, nil
	}

	f, err := src.Provider.Open(src.Path)
	if err != nil {
		return nil, err
	}
	defer safe.Close(f, unsafe.Ignore)

	raw, err := unpackSnapshot(f)
	if err != nil {
		return nil, fmt.Errorf("read snapshot %q: %w", src.Path, err)
	}

	// A snapshot is keyed by days, at midnight UTC, as the heat map
	// requires; the same instant written with a zero offset is the same day.
	data := make(HeatMap, len(raw))
	for _, entry := range raw {
		ts, err := time.Parse(time.RFC3339, entry.key)
		if err != nil {
			return nil, fmt.Errorf("read snapshot %q: key %q is not an RFC 3339 time", src.Path, entry.key)
		}
		count := entry.count
		day := ts.UTC()
		if !day.Equal(xtime.TruncateToDay(day)) {
			return nil, fmt.Errorf("read snapshot %q: key %q is not a day at midnight UTC",
				src.Path, ts.Format(time.RFC3339Nano))
		}
		if _, present := data[day]; present {
			return nil, fmt.Errorf("read snapshot %q: key %q repeats the day %s",
				src.Path, ts.Format(time.RFC3339Nano), day.Format(xtime.DateOnly))
		}
		data.SetCount(day, count)
	}
	src.data = data
	return src.data, nil
}

func (src *fileSource) Store(chm HeatMap) error {
	f, err := src.Provider.Create(src.Path)
	if err != nil {
		return err
	}
	defer safe.Close(f, unsafe.Ignore)

	src.data = chm
	return packer.Pack(f, src.data)
}

func NewUpstreamSource(src Contributor, year time.Time) *upstreamSource {
	return &upstreamSource{
		Provider: src,
		Year:     year,
		Now:      time.Now,
	}
}

type upstreamSource struct {
	Provider Contributor
	Year     time.Time
	Now      func() time.Time

	data HeatMap
}

func (src upstreamSource) Location() string {
	return fmt.Sprintf("upstream:year(%d)", src.Year.Year())
}

func (src *upstreamSource) Fetch(ctx context.Context) (HeatMap, error) {
	if src.data != nil {
		return src.data, nil
	}

	if src.Year.Year() < MinYear {
		return nil, fmt.Errorf("fetch %s: the year is before %d", src.Location(), MinYear)
	}
	scope, err := xtime.RangeByYears(src.Year, 0, false).ExcludeFuture(src.Now())
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", src.Location(), err)
	}
	src.data, err = src.Provider.ContributionHeatMap(ctx, scope)
	return src.data, err
}

// snapshotEntry is a raw pair of a snapshot, before its key is checked.
type snapshotEntry struct {
	key   string
	count uint
}

// unpackSnapshot strictly reads a snapshot: exactly one top-level object
// keyed by strings, without duplicate keys, with counts that are integers
// from 0 to math.MaxInt64, so the signed difference of two counts
// never overflows. It bypasses the packer to control the decoding.
func unpackSnapshot(file io.Input) ([]snapshotEntry, error) {
	switch ext := strings.ToLower(filepath.Ext(file.Name())); ext {
	case ".json":
		return unpackJSONSnapshot(file)
	case ".yml", ".yaml":
		return unpackYAMLSnapshot(file)
	default:
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}
}

func unpackJSONSnapshot(r io.Reader) ([]snapshotEntry, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil {
		if errors.Is(err, stdio.EOF) {
			return nil, errors.New("the snapshot is empty")
		}
		return nil, err
	}
	if delim, is := tok.(json.Delim); !is || delim != '{' {
		if tok == nil {
			tok = "null"
		}
		return nil, fmt.Errorf("the snapshot is not an object: %v", tok)
	}

	var entries []snapshotEntry
	seen := make(map[string]struct{})
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, is := tok.(string)
		if !is {
			return nil, fmt.Errorf("unexpected token %v", tok)
		}
		if _, present := seen[key]; present {
			return nil, fmt.Errorf("key %q is duplicated", key)
		}
		seen[key] = struct{}{}

		var value any
		if err := dec.Decode(&value); err != nil {
			return nil, err
		}
		count, err := snapshotCount(value)
		if err != nil {
			return nil, fmt.Errorf("key %q: %w", key, err)
		}
		entries = append(entries, snapshotEntry{key, count})
	}
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	if tok, err := dec.Token(); !errors.Is(err, stdio.EOF) {
		if err != nil {
			return nil, fmt.Errorf("unexpected data after the object: %w", err)
		}
		return nil, fmt.Errorf("unexpected data after the object: %v", tok)
	}
	return entries, nil
}

func unpackYAMLSnapshot(r io.Reader) ([]snapshotEntry, error) {
	// yaml.v2 rejects duplicate keys only in the strict mode
	dec := yaml.NewDecoder(r)
	dec.SetStrict(true)

	var doc any
	if err := dec.Decode(&doc); err != nil {
		if errors.Is(err, stdio.EOF) {
			return nil, errors.New("the snapshot is empty")
		}
		return nil, err
	}
	obj, is := doc.(map[any]any)
	if !is {
		return nil, fmt.Errorf("the snapshot is not a mapping: %T", doc)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, stdio.EOF) {
		if err != nil {
			return nil, fmt.Errorf("unexpected data after the mapping: %w", err)
		}
		return nil, errors.New("unexpected document after the first one")
	}

	entries := make([]snapshotEntry, 0, len(obj))
	for rawKey, value := range obj {
		var key string
		switch k := rawKey.(type) {
		case string:
			key = k
		case time.Time:
			key = k.Format(time.RFC3339Nano)
		default:
			return nil, fmt.Errorf("key %v is not a string", rawKey)
		}
		count, err := snapshotCount(value)
		if err != nil {
			return nil, fmt.Errorf("key %q: %w", key, err)
		}
		entries = append(entries, snapshotEntry{key, count})
	}
	// the map order is random, so sort the keys to report errors stably
	sort.Slice(entries, func(i, j int) bool { return entries[i].key < entries[j].key })
	return entries, nil
}

// snapshotCount accepts only an integer from 0 to math.MaxInt64.
func snapshotCount(value any) (uint, error) {
	invalid := func() error {
		return fmt.Errorf("the count %#v is not an integer from 0 to %d", value, int64(math.MaxInt64))
	}
	switch v := value.(type) {
	case json.Number:
		n, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil || n < 0 {
			return 0, invalid()
		}
		return uint(n), nil
	case int:
		if v < 0 {
			return 0, invalid()
		}
		return uint(v), nil
	case int64:
		if v < 0 {
			return 0, invalid()
		}
		return uint(v), nil
	case uint64:
		if v > math.MaxInt64 {
			return 0, invalid()
		}
		return uint(v), nil
	default:
		return 0, invalid()
	}
}
