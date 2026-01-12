package contribution

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/assert"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

// maxWeeks limits the span of weeks around a date to a year:
// a longer one fits no table and only inflates the requests.
const maxWeeks = 53

// rfc3339 is the profile of RFC 3339 a date may have: two-digit fields,
// an optional fraction after a dot, and Z or a ±HH:MM offset.
// The ranges of the date and time are checked by time.Parse.
var rfc3339 = regexp.MustCompile(
	`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-](\d{2}):(\d{2}))$`,
)

// tokenNote tells the commands that need GitHub data about the token.
const tokenNote = `
The contribution calendar is read through the GitHub GraphQL API for
the owner of the token, which is required even for a public profile:
set GITHUB_TOKEN or pass --token.`

// ErrNoToken reports a command that needs GitHub data, but has no token.
var ErrNoToken = errors.New(
	"a GitHub token is required: the contribution calendar is read through the GitHub GraphQL API, " +
		"which needs one even for a public profile; please set GITHUB_TOKEN or pass --token",
)

// RequireToken declines a command that needs GitHub data before any request,
// if there is no token to authorize it.
func RequireToken(cnf *config.Tool) error {
	if cnf.Token == "" {
		return ErrNoToken
	}
	return nil
}

func Datetime(t, now time.Time) string {
	now = now.In(t.Location())
	sign := "-"
	if t.After(now) {
		sign = "+"
	}

	days := t.Sub(now) / xtime.Day
	if days < 0 {
		days = -days
	}
	tail := t.Sub(now) % xtime.Day
	if tail < 0 {
		tail = -tail
	}
	normalized := strings.ToUpper(tail.Truncate(time.Second).String())

	if days > 0 {
		return fmt.Sprintf("%s%dd%s", sign, days, normalized)
	}
	return fmt.Sprintf("%s%s", sign, normalized)
}

// FallbackDate returns the anchor for an empty or git date: the author date
// of HEAD in the repository GIT_DIR points to, which must open, or else in
// the one around the working directory. Outside a repository, and in
// a repository without commits, the anchor is now. For any other date
// the value is not used, and it is now.
func FallbackDate(args []string, now time.Time) (time.Time, error) {
	if len(args) > 0 {
		raw := strings.Split(args[0], "/")
		rawDate := raw[0]
		if rawDate != "" && rawDate != "git" {
			return now, nil
		}
	}

	var (
		repo *git.Repository
		err  error
	)
	if dir := os.Getenv("GIT_DIR"); dir != "" {
		// exactly the specified repository, as git does, without a search
		repo, err = git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
		if err != nil {
			return time.Time{}, fmt.Errorf("open the repository GIT_DIR=%q points to: %w", dir, err)
		}
	} else {
		repo, err = git.PlainOpenWithOptions("", &git.PlainOpenOptions{
			DetectDotGit:          true,
			EnableDotGitCommonDir: true, // a linked worktree keeps its refs there
		})
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return now, nil
		}
		if err != nil {
			return time.Time{}, fmt.Errorf("open the repository: %w", err)
		}
	}

	head, err := repo.Head()
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return now, nil // there are no commits yet
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("resolve HEAD: %w", err)
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return time.Time{}, fmt.Errorf("read the HEAD commit %s: %w", head.Hash(), err)
	}
	return commit.Author.When, nil
}

// ParseDate parses the argument in format date[/weeks], where date is empty,
// git, now, YYYY, YYYY-MM, YYYY-MM-DD, or RFC 3339, and weeks are N for
// a span centered on the date, +N for weeks ahead, or -N for weeks behind.
// An empty or git date is defaultDate, now is now, and no weeks are
// defaultWeeks, centered if positive, as N is.
func ParseDate(
	args []string,
	defaultDate time.Time,
	defaultWeeks int,
	now time.Time,
) (contribution.DateOptions, error) {
	// trick to skip length check
	args = append(args, "")

	var (
		opts contribution.DateOptions
		err  error
	)
	var rawDate, rawWeeks string
	raw := strings.Split(args[0], "/")
	switch len(raw) {
	case 2:
		rawDate, rawWeeks = raw[0], raw[1]
	case 1:
		rawDate, rawWeeks = raw[0], ""
	default:
		return opts, fmt.Errorf("too many parts")
	}

	var date time.Time
	switch l := len(rawDate); {
	case rawDate == "" || rawDate == "git":
		date = defaultDate
	case rawDate == "now":
		date = now
	case l == len(xtime.YearOnly):
		date, err = time.Parse(xtime.YearOnly, rawDate)
	case l == len(xtime.YearAndMonth):
		date, err = time.Parse(xtime.YearAndMonth, rawDate)
	case l == len(xtime.DateOnly):
		date, err = time.Parse(xtime.DateOnly, rawDate)
	default:
		// time.Parse is lenient to a single-digit hour, a comma before
		// the fraction, and an offset out of range, so check the syntax first
		date, err = parseRFC3339(rawDate)
	}
	if err == nil && date.UTC().Year() < contribution.MinYear {
		err = fmt.Errorf("the year is before %d", contribution.MinYear)
	}
	if err != nil {
		return opts, fmt.Errorf("parse date %q: %w", rawDate, err)
	}
	opts.Value = date

	var weeks = defaultWeeks
	if len(raw) == 2 && rawWeeks == "" {
		return opts, fmt.Errorf("parse weeks %q: nothing after the slash", rawWeeks)
	}
	if rawWeeks != "" {
		weeks, err = strconv.Atoi(rawWeeks)
		if err != nil {
			return opts, fmt.Errorf("parse weeks %q: %w", rawWeeks, err)
		}
		if weeks < -maxWeeks || weeks > maxWeeks {
			return opts, fmt.Errorf("parse weeks %q: the span is longer than %d weeks", rawWeeks, maxWeeks)
		}
		// +%d and positive %d have the same value, but different semantic
		// invariant: len(rawWeeks) > 0, because weeks > 0
		opts.Half = weeks > 0 && rawWeeks[0] != '+'
	} else {
		// the default span is centered if positive, the same as N
		opts.Half = weeks > 0
	}
	opts.Weeks = weeks

	return opts, nil
}

// parseRFC3339 parses the date in the supported profile of RFC 3339.
func parseRFC3339(raw string) (time.Time, error) {
	match := rfc3339.FindStringSubmatch(raw)
	if match == nil {
		return time.Time{}, fmt.Errorf("not in format %s", "YYYY-MM-DDTHH:MM:SS[.F](Z|±HH:MM)")
	}
	if match[1] != "" {
		hours, _ := strconv.Atoi(match[1])
		minutes, _ := strconv.Atoi(match[2])
		if hours > 23 || minutes > 59 {
			return time.Time{}, fmt.Errorf("the offset %s:%s is out of range", match[1], match[2])
		}
	}
	return time.Parse(time.RFC3339, raw)
}

func TableView(
	cmd *cobra.Command,
	heats contribution.HeatMap,
	scope xtime.Range,
	opts ...func(time.Time, string) string,
) {
	assert.True(func() bool { return scope.From().Weekday() == time.Sunday })

	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "Day / Week"},
		},
	}
	var weeks int
	for i := scope.From(); i.Before(scope.To()); i = i.Add(xtime.Week) {
		_, week := xtime.GregorianWeek(i) // weeks start on Sunday, as the rows do
		table.Header.Cells = append(table.Header.Cells, &simpletable.Cell{
			Align: simpletable.AlignCenter,
			Text:  fmt.Sprintf("#%02d", week),
		})
		weeks++
	}
	table.Header.Cells = append(table.Header.Cells, &simpletable.Cell{
		Align: simpletable.AlignCenter,
		Text:  "Date",
	})

	shown := make(contribution.HeatMap)
	for i, cursor := time.Sunday, scope.From(); i <= time.Saturday; i++ {
		row := append(make([]*simpletable.Cell, 0, weeks+1), &simpletable.Cell{Text: i.String()})
		for j := 0; j < weeks; j++ {
			cell := cursor.Add(time.Duration(j) * xtime.Week)

			count := heats.Count(cell)
			if !cell.After(scope.To()) {
				shown.SetCount(cell, count)
			}
			text := "-"
			if count > 0 {
				text = strconv.FormatUint(uint64(count), 10)
			} else if cell.After(scope.To()) {
				text = "?"
			}
			for _, opt := range opts {
				text = opt(cell, text)
			}

			row = append(row, &simpletable.Cell{Align: simpletable.AlignCenter, Text: text})
			if j+1 == weeks {
				row = append(row, &simpletable.Cell{
					Align: simpletable.AlignCenter,
					Text:  cell.Format(xtime.DayStamp),
				})
			}
		}
		cursor = cursor.Add(xtime.Day)
		table.Body.Cells = append(table.Body.Cells, row)
	}

	table.Footer = &simpletable.Footer{
		Cells: []*simpletable.Cell{
			{
				Align: simpletable.AlignRight,
				Span:  len(table.Header.Cells),
				Text:  distribution(shown),
			},
		},
	}
	table.SetStyle(simpletable.StyleCompactLite)
	cmd.PrintErrln("\n" + table.String() + "\n")
}

// distribution tells how many days have each contribution count,
// e.g., distribution{5: 123, 10: 23, 15: 2}.
func distribution(heats contribution.HeatMap) string {
	rows := contribution.HistogramByCount(heats, contribution.OrderByCount)
	pairs := make([]string, 0, len(rows))
	for _, row := range rows {
		pairs = append(pairs, fmt.Sprintf("%d: %d", row.Count, row.Frequency))
	}
	return "distribution{" + strings.Join(pairs, ", ") + "}"
}
