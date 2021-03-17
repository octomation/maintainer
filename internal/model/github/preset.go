package github

type Preset map[string][]Label

var preset = Preset{
	"default": []Label{
		{
			Name:  "bug",
			Color: "d73a4a",
			Desc:  "Something isn't working",
		},
		{
			Name:  "documentation",
			Color: "0075ca",
			Desc:  "Improvements or additions to documentation",
		},
		{
			Name:  "duplicate",
			Color: "cfd3d7",
			Desc:  "This issue or pull request already exists",
		},
		{
			Name:  "enhancement",
			Color: "a2eeef",
			Desc:  "New feature or request",
		},
		{
			Name:  "good first issue",
			Color: "7057ff",
			Desc:  "Good for newcomers",
		},
		{
			Name:  "help wanted",
			Color: "008672",
			Desc:  "Extra attention is needed",
		},
		{
			Name:  "invalid",
			Color: "e4e669",
			Desc:  "This doesn't seem right",
		},
		{
			Name:  "question",
			Color: "d876e3",
			Desc:  "Further information is requested",
		},
		{
			Name:  "wontfix",
			Color: "ffffff",
			Desc:  "This will not be worked on",
		},
	},
	"octolab": []Label{
		{
			Name:  "kind: bug",
			Color: "e63946",
			Desc:  "New bug report.",
		},
		{
			Name:  "kind: feature",
			Color: "1d3557",
			Desc:  "New feature request.",
		},
		{
			Name:  "kind: improvement",
			Color: "a8dadc",
			Desc:  "New improvement proposal.",
		},
		{
			Name:  "scope: code",
			Color: "5c6b73",
			Desc:  "Issue related to source code.",
		},
		{
			Name:  "scope: docs",
			Color: "9db4c0",
			Desc:  "Issue related to documentation.",
		},
		{
			Name:  "scope: test",
			Color: "c2dfe3",
			Desc:  "Issue related to tests.",
		},
		{
			Name:  "scope: eqpt",
			Color: "e0fbfc",
			Desc:  "Issue related to auxiliary code, e.g. CI config, Makefiles, etc.",
		},
		{
			Name:  "level: critical",
			Color: "f25c54",
			Desc:  "Issue has critical severity and needs to be fixed as soon as possible.",
		},
		{
			Name:  "level: normal",
			Color: "f4845f",
			Desc:  "Issue has normal severity and needs to be fixed in the nearest iteration.",
		},
		{
			Name:  "level: low",
			Color: "f7b267",
			Desc:  "Issue has low severity and should be fixed when possible.",
		},
		{
			Name:  "difficulty: easy",
			Color: "f7d1cd",
			Desc:  "Issue is easy to implement.",
		},
		{
			Name:  "difficulty: medium",
			Color: "d1b3c4",
			Desc:  "Issue has medium complexity.",
		},
		{
			Name:  "difficulty: hard",
			Color: "735d78",
			Desc:  "Issue is hard to implement or reproduce.",
		},
		{
			Name:  "good first issue",
			Color: "90be6d",
			Desc:  "Good for newcomers.",
		},
		{
			Name:  "help wanted",
			Color: "577590",
			Desc:  "Extra attention is needed.",
		},
		{
			Name:  "invalid",
			Color: "f94144",
			Desc:  "This doesn't seem right.",
		},
		{
			Name:  "wontfix",
			Color: "f9c74f",
			Desc:  "This will not be worked on.",
		},
		{
			Name:  "duplicate",
			Color: "f8961e",
			Desc:  "This issue or pull request already exists.",
		},
		{
			Name:  "question",
			Color: "43aa8b",
			Desc:  "Further information is requested.",
		},
	},
}
