package github

// Label represents a GitHub label.
type Label struct {
	ID    int64
	Name  string
	Color string
	Desc  string
}
