package github

import (
	"context"
	"embed"
	"path/filepath"

	"github.com/google/go-github/v44/github"
	"go.octolab.org/pointer"
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

// UpdateLabels updates labels in GitHub.
func (srv *service) UpdateLabels(
	ctx context.Context,
	src model.Remote,
	set model.LabelSet,
) error {
	current, err := srv.Labels(ctx, src)
	if err != nil {
		return err
	}

	owner, repo := src.OwnerAndName()

	for _, label := range set.Labels {
		if label.IsNew() {
			_, _, err := srv.client.Issues.CreateLabel(ctx, owner, repo, &github.Label{
				Name:        pointer.ToString(label.Name),
				Color:       pointer.ToString(label.Color),
				Description: pointer.ToString(label.Desc),
			})
			if err != nil {
				return err // TODO:accumulate
			}
			continue
		}

		// TODO:refactor abstraction leak
		prev := current.FindByID(label.ID)
		if prev == nil {
			continue
		}

		if label.IsEmpty() {
			_, err := srv.client.Issues.DeleteLabel(ctx, owner, repo, prev.Name)
			if err != nil {
				return err // TODO:accumulate
			}
			continue
		}

		if label.IsChanged(*prev) {
			dto := new(github.Label)
			dto.ID = pointer.ToInt64(label.ID)
			dto.Name = pointer.ToString(label.Name)
			dto.Color = pointer.ToString(label.Color)
			dto.Description = pointer.ToString(label.Desc)

			_, _, err := srv.client.Issues.EditLabel(ctx, owner, repo, prev.Name, dto)
			if err != nil {
				return err // TODO:accumulate
			}
			continue
		}
	}

	return nil
}
