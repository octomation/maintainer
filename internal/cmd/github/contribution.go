package github

import (
	"time"

	"github.com/spf13/cobra"
)

//
// $ maintainer github contribution suggest --since=2021-01-01
//
// https://github.com/kamilsk?tab=overview&from=2021-12-01&to=2021-12-31
//
// $('.js-calendar-graph-svg rect.ContributionCalendar-day')
//   data-date
//   data-level
//

func Contribution(github GitHub) *cobra.Command {
	cmd := cobra.Command{
		Use: "contribution",
	}

	suggest := cobra.Command{
		Use: "suggest",
		RunE: func(cmd *cobra.Command, args []string) error {
			chm, err := github.ContributionHeatMap(cmd.Context(), time.Now())
			if err != nil {
				return err
			}

			cmd.Println("-123d", chm)
			return nil
		},
	}
	cmd.AddCommand(&suggest)

	return &cmd
}
