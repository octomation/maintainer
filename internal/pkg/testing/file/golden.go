package file

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func Golden(name string, data any, fake bool) {
	fs := afero.NewOsFs()
	if fake {
		fs = afero.NewMemMapFs()
	}

	file, err := fs.Create(name)
	if err != nil {
		panic(err)
	}

	if err := pack(file, data); err != nil {
		panic(err)
	}
}

func Restore(name string, ptr interface{}) {
	fs := afero.NewOsFs()
	file, err := fs.Open(name)
	if err != nil {
		panic(err)
	}

	if err := unpack(file, ptr); err != nil {
		panic(err)
	}
}

func pack(file afero.File, data any) error {
	format := strings.ToLower(filepath.Ext(file.Name()))

	switch format {
	case ".json":
		return json.NewEncoder(file).Encode(data)
	case ".yml", ".yaml":
		return yaml.NewEncoder(file).Encode(data)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func unpack(file afero.File, ptr any) error {
	format := strings.ToLower(filepath.Ext(file.Name()))

	switch format {
	case ".json":
		return json.NewDecoder(file).Decode(ptr)
	case ".yml", ".yaml":
		return yaml.NewDecoder(file).Decode(ptr)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}
