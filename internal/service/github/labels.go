package github

import (
	"context"
	"embed"

	"github.com/google/go-github/v33/github"

	model "go.octolab.org/toolset/maintainer/internal/model/github"
)

//go:embed preset/*.yml
var presets embed.FS

// Labels lists all labels for a repository.
func (srv *service) Labels(
	ctx context.Context,
	src model.Remote,
) (*model.LabelSet, error) {
	owner, repo := src.OwnerAndName()
	opt := new(github.ListOptions)
	list, _, err := srv.client.Issues.ListLabels(ctx, owner, repo, opt)
	if err != nil {
		return nil, err
	}

	result := new(model.LabelSet)
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
