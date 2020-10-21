package github

import (
	"context"
	"sort"
	"sync"

	"github.com/google/go-github/v33/github"
	"golang.org/x/sync/errgroup"

	entity "go.octolab.org/toolset/maintainer/internal/entity/github"
	"go.octolab.org/toolset/maintainer/internal/errors"
)

func (manager *manager) CompareLabels(
	ctx context.Context,
	expected,
	obtained entity.RepositoryURN,
) ([]entity.LabelTransform, error) {
	repositories, err := manager.RepositoryWithLabels(ctx, expected, obtained)
	if err != nil {
		return nil, err
	}

	for _, repository := range repositories {
		sort.Slice(repository.Labels, func(i, j int) bool {
			return repository.Labels[i].Name < repository.Labels[j].Name
		})
	}

	src, dst := repositories[0], repositories[1]
	_, _ = src, dst
	return nil, nil
}

func (manager *manager) RepositoryWithLabels(
	ctx context.Context,
	list ...entity.RepositoryURN,
) ([]entity.Repository, error) {
	result := make([]entity.Repository, 0, len(list))

	for _, urn := range list {
		var (
			mutex      sync.Mutex
			repository entity.Repository
		)

		owner, repo := urn.Split()
		group, ctx := errgroup.WithContext(ctx)

		group.Go(func() error {
			data, _, err := manager.client.Repositories.Get(ctx, owner, repo)
			if err != nil {
				return err
			}

			mutex.Lock()
			repository.ID = data.GetID()
			repository.URN = entity.RepositoryURN(data.GetFullName())
			mutex.Unlock()
			return nil
		})

		group.Go(func() error {
			opt := &github.ListOptions{}
			data, _, err := manager.client.Issues.ListLabels(ctx, owner, repo, opt)
			if err != nil {
				return err
			}
			if len(data) == 0 {
				return nil
			}

			labels := make([]entity.Label, 0, len(data))
			for _, label := range data {
				labels = append(labels, entity.Label{
					ID:    label.GetID(),
					Name:  label.GetName(),
					Color: label.GetColor(),
					Desc:  label.GetDescription(),
				})
			}

			mutex.Lock()
			repository.Labels = labels
			mutex.Unlock()
			return nil
		})

		if err := group.Wait(); err != nil {
			return nil, err
		}
		result = append(result, repository)
	}

	if len(result) != len(list) {
		return nil, errors.Inconsistent
	}
	return result, nil
}

var _ = map[string]struct {
	color string
	desc  string
	to    entity.Label
}{
	"bug": {
		color: "d73a4a",
		desc:  "Something isn't working",
		to: entity.Label{
			Name:  "",
			Color: "",
			Desc:  "",
		},
	},
	"documentation": {
		color: "0075ca",
		desc:  "Improvements or additions to documentation",
		to: entity.Label{
			Name:  "",
			Color: "",
			Desc:  "",
		},
	},
	"duplicate": {
		color: "cfd3d7",
		desc:  "This issue or pull request already exists",
		to: entity.Label{
			Name:  "",
			Color: "",
			Desc:  "",
		},
	},
	"enhancement": {
		color: "a2eeef",
		desc:  "New feature or request",
		to: entity.Label{
			Name:  "",
			Color: "",
			Desc:  "",
		},
	},
	"good first issue": {
		color: "7057ff",
		desc:  "Good for newcomers",
		to: entity.Label{
			Name:  "",
			Color: "",
			Desc:  "",
		},
	},
	"help wanted": {
		color: "008672",
		desc:  "Extra attention is needed",
		to: entity.Label{
			Name:  "invalid",
			Color: "e4e669",
			Desc:  "This doesn't seem right",
		},
	},
	"invalid": {
		color: "e4e669",
		desc:  "This doesn't seem right",
		to: entity.Label{
			Name:  "",
			Color: "",
			Desc:  "",
		},
	},
	"question": {
		color: "d876e3",
		desc:  "Further information is requested",
		to: entity.Label{
			Name:  "",
			Color: "",
			Desc:  "",
		},
	},
	"wontfix": {
		color: "ffffff",
		desc:  "This will not be worked on",
		to: entity.Label{
			Name:  "",
			Color: "",
			Desc:  "",
		},
	},
}
