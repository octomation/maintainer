package labels

import (
	"context"
	"fmt"
	"strconv"

	"github.com/alexeyco/simpletable"
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/entity"
)

func NewListCommand(provider Provider) *cobra.Command {
	command := cobra.Command{
		Use:  "list",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()

			list := entity.ConvertStringsToRepositoryURNs(args)
			repositories, err := provider.RepositoryWithLabels(ctx, list...)
			if err != nil {
				return err
			}

			for _, repository := range repositories {
				cmd.Println("## " + repository.URN.String())
				table := simpletable.New()
				table.Header = &simpletable.Header{
					Cells: []*simpletable.Cell{
						{Text: "ID"},
						{Text: "Name"},
						{Text: "Color"},
						{Text: "Description"},
					},
				}
				for _, label := range repository.Labels {
					table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
						{Text: strconv.FormatInt(label.ID, 10)},
						{Text: label.Name},
						{Text: label.Color},
						{Text: label.Desc},
					})
				}
				table.Footer = &simpletable.Footer{
					Cells: []*simpletable.Cell{
						{Text: fmt.Sprintf("Total: %d", len(repository.Labels)), Span: 4},
					},
				}
				table.SetStyle(simpletable.StyleDefault)
				cmd.Println(table.String())
			}
			return nil
		},
	}

	return &command
}
