package github

import (
	"context"

	"github.com/google/go-github/v33/github"

	model "go.octolab.org/toolset/maintainer/internal/model/github"
)

// Labels lists all labels for a repository.
func (srv *service) Labels(
	ctx context.Context,
	owner, repo string,
) ([]model.Label, error) {
	opt := new(github.ListOptions)
	list, _, err := srv.client.Issues.ListLabels(ctx, owner, repo, opt)
	if err != nil {
		return nil, err
	}

	result := make([]model.Label, 0, len(list))
	for _, dto := range list {
		result = append(result, model.Label{
			ID:    dto.GetID(),
			Name:  dto.GetName(),
			Color: dto.GetColor(),
			Desc:  dto.GetDescription(),
		})
	}
	return result, nil
}
