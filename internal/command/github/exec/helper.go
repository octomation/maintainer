package exec

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"

	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func FallbackDate(args []string) time.Time {
	fallback := time.Now().UTC()
	if len(args) > 0 {
		return fallback
	}

	repo, err := git.PlainOpenWithOptions("", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return fallback
	}
	head, err := repo.Head()
	if err != nil {
		return fallback
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return fallback
	}
	return commit.Author.When.UTC()
}

func ParseDate(
	args []string,
	defaultDate time.Time,
	defaultWeeks int,
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
	case rawDate == "":
		date = defaultDate
	case rawDate == "now":
		date = time.Now()
	case l == len(xtime.RFC3339Year):
		date, err = time.Parse(xtime.RFC3339Year, rawDate)
	case l == len(xtime.RFC3339Month):
		date, err = time.Parse(xtime.RFC3339Month, rawDate)
	case l == len(xtime.RFC3339Day):
		date, err = time.Parse(xtime.RFC3339Day, rawDate)
	case l == 20 || l == len(time.RFC3339):
		date, err = time.Parse(time.RFC3339, rawDate)
	default:
		err = fmt.Errorf("unsupported format")
	}
	if err != nil {
		return opts, fmt.Errorf("parse date %q: %w", rawDate, err)
	}
	opts.Value = date

	var weeks = defaultWeeks
	if rawWeeks != "" {
		weeks, err = strconv.Atoi(rawWeeks)
		if err != nil {
			return opts, fmt.Errorf("parse weeks %q: %w", rawWeeks, err)
		}
		// +%d and positive %d have the same value, but different semantic
		// invariant: len(rawWeeks) > 0, because weeks > 0
		if weeks > 0 && rawWeeks[0] != '+' {
			opts.Half = true
		}
	} else {
		opts.Half = true
	}
	opts.Weeks = weeks

	return opts, nil
}
