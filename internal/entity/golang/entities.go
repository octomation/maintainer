package golang

type Module struct {
	Name     string   `yaml:"prefix"`
	Import   []Import `yaml:"import"`
	Packages []string `yaml:"packages"`
	Tags     []string `yaml:"tags"`
}

type Import struct {
	URL    string `yaml:"url"`
	VCS    string `yaml:"vcs"`
	Source Source `yaml:"source"`
}

type Source struct {
	URL  string `yaml:"url"`
	Dir  string `yaml:"dir"`
	File string `yaml:"file"`
}
