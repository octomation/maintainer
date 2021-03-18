package github

import (
	"context"
	"embed"
	"regexp"

	"github.com/google/go-github/v33/github"

	model "go.octolab.org/toolset/maintainer/internal/model/github"
)

var (
	skipOp    = regexp.MustCompile(`^skip$`)
	deleteOp  = regexp.MustCompile(`^delete$`)
	replaceOp = regexp.MustCompile(`^replace\([^()]+\)$`)
)

//go:embed preset/*.yml
var presets embed.FS

// Labels lists all labels for a repository.
func (srv *service) Labels(
	ctx context.Context,
	src model.GitHub,
) ([]model.Label, error) {
	owner, repo := src.OwnerAndName()
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
