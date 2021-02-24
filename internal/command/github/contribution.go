package github

import (
	"time"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/command/github/view"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
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

	lookup := cobra.Command{
		Use: "lookup",
		RunE: func(cmd *cobra.Command, args []string) error {
			weeks := 7
			date, err := time.Parse(xtime.RFC3339Day, "2021-02-24")
			if err != nil {
				return err
			}

			chm, err := github.ContributionHeatMap(cmd.Context(), date)
			if err != nil {
				return err
			}

			r := xtime.RangeByWeeks(date, weeks).TrimByYear(date.Year())

			histogram := contribution.HistogramByWeekday(chm.Subset(r.From(), r.To()), false)
			report := make([]view.WeekReport, 0, 4)

			// TODO:refactoring normalize for Sunday as a first day of week and ISO week
			_, start := r.From().ISOWeek()
			if r.From().Weekday() == time.Sunday {
				start++
			}
			for i := r.From(); i.Before(r.To()); i = i.Add(xtime.Day) {
				weekday := i.Weekday()
				_, current := i.ISOWeek()
				if weekday == time.Sunday {
					current++
				}

				idx := current % start
				if len(report) < idx+1 {
					report = append(report, view.WeekReport{
						Number: current,
						Report: make(map[time.Weekday]int),
					})
				}

				var count int
				if len(histogram) > 0 {
					row := histogram[0]
					if row.Day.Equal(i) {
						histogram = histogram[1:]
						count = row.Sum
					}
				}
				report[idx].Report[weekday] = count
			}

			return view.Lookup(r, report, cmd)
		},
	}
	cmd.AddCommand(&lookup)

	return &cmd
}
