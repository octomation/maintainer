package contribution

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"
	"gopkg.in/yaml.v2"

	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

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

	file, err := src.Provider.Open(src.Path)
	if err != nil {
		return nil, err
	}
	defer safe.Close(file, unsafe.Ignore)

	var data HeatMap
	format := strings.ToLower(filepath.Ext(file.Name()))
	switch format {
	case ".json":
		err := json.NewDecoder(file).Decode(&data)
		src.data = data
		return data, err
	case ".yml", ".yaml":
		err := yaml.NewDecoder(file).Decode(&data)
		src.data = data
		return data, err
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

type UpstreamSource struct {
	Provider Contributor
	Year     time.Time

	data HeatMap
}

func (src UpstreamSource) Location() string {
	return fmt.Sprintf("upstream:year(%s)", src.Year.Format(xtime.RFC3339Year))
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
