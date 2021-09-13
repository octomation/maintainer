package contribution

import (
	"context"
	"encoding/json"
	"fmt"
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

type FileSource struct {
	Provider afero.Fs
	Path     string

	data HeatMap
}

func (src FileSource) Location() string {
	return fmt.Sprintf("file:%s", src.Path)
}

func (src *FileSource) Fetch(_ context.Context) (HeatMap, error) {
	if src.data != nil {
		return src.data, nil
	}

	f, err := src.Provider.Open(src.Path)
	if err != nil {
		return nil, err
	}
	defer safe.Close(f, unsafe.Ignore)

	if err := packer.Unpack(f, &src.data); err != nil {
		return nil, err
	}
	return src.data, nil
}

type UpstreamSource struct {
	Provider Contributor
	Year     time.Time

	data HeatMap
}

func (src UpstreamSource) Location() string {
	return fmt.Sprintf("upstream:year(%d)", src.Year.Year())
}

func (src *UpstreamSource) Fetch(ctx context.Context) (HeatMap, error) {
	if src.data != nil {
		return src.data, nil
	}

	var err error
	scope := xtime.RangeByYears(src.Year, 0, false).ExcludeFuture()
	src.data, err = src.Provider.ContributionHeatMap(ctx, scope)
	return src.data, err
}
