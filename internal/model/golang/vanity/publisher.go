package vanity

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	xstrings "go.octolab.org/strings"

	"go.octolab.org/toolset/maintainer/internal/model/golang"
)

func New(host string, fs afero.Fs) *publisher {
	return &publisher{host, fs}
}

type publisher struct {
	host string
	fs   afero.Fs
}

func (publisher *publisher) PublishAt(root string, modules []golang.Module) error {
	for _, module := range modules {
		if len(module.Import) == 0 {
			continue
		}

		prefix := xstrings.FirstNotEmpty(module.Prefix, module.Name)
		for _, pkg := range fill(prefix, module.Packages) {
			dir := filepath.Join(append(
				[]string{root},
				strings.Split(strings.TrimPrefix(pkg, publisher.host), "/")...,
			)...)
			if err := publisher.fs.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
			file, err := publisher.fs.Create(filepath.Join(dir, "index.html"))
			if err != nil {
				return err
			}

			// TODO:rfc:#3 add support multiple source
			module := module
			for _, imports := range module.Import[:1] {
				if err := tpl.Execute(file, Meta{
					Package: pkg,
					Import: MetaImport{
						Prefix:   prefix,
						VCS:      imports.VCS,
						RepoRoot: imports.URL,
					},
					Source: MetaSource{
						URL:  imports.Source.URL,
						Dir:  imports.Source.Dir,
						File: imports.Source.File,
					},
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
