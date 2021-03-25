package github

import (
	"context"
	"embed"
	"path/filepath"

	"github.com/google/go-github/v33/github"
	"gopkg.in/yaml.v2"

	model "go.octolab.org/toolset/maintainer/internal/model/github"
)

//go:embed preset/*.yml
var presets embed.FS

// Labels lists all labels for a repository.
func (srv *service) Labels(
	ctx context.Context,
	src model.Remote,
) (model.LabelSet, error) {
	var result model.LabelSet

	owner, repo := src.OwnerAndName()
	opt := new(github.ListOptions)
	opt.Page, opt.PerPage = 1, 100
	// assumption: all labels can be fetched by one request
	list, _, err := srv.client.Issues.ListLabels(ctx, owner, repo, opt)
	if err != nil {
		return result, err
	}

	result.Name = src.ID()
	result.Labels = make([]model.Label, 0, len(list))
	for _, dto := range list {
		result.Labels = append(result.Labels, model.Label{
			ID:    dto.GetID(),
			Name:  dto.GetName(),
			Color: dto.GetColor(),
			Desc:  dto.GetDescription(),
		})
	}
	return result, nil
}

// PatchLabels updates labels with the specified preset.
func (srv *service) PatchLabels(
	_ context.Context,
	current model.LabelSet,
	name string,
) (model.LabelSet, error) {
	f, err := presets.Open(filepath.Join("preset", name+".yml"))
	if err != nil {
		return current, err
	}

	var preset model.LabelPreset
	if err := yaml.NewDecoder(f).Decode(&preset); err != nil {
		return current, err
	}

	// to delete and update
	for i, label := range current.Labels {
		diff := preset.ExtractMatched(label)
		current.Labels[i].Apply(diff)
	}
	// to insert
	for _, patch := range preset.Labels {
		current.Labels = append(current.Labels, patch.Label)
	}

	return current, nil
}
