package github

// Label represents a GitHub label.
type Label struct {
	ID    int64  `yaml:",omitempty"`
	Name  string `yaml:",omitempty"`
	Color string `yaml:",omitempty"`
	Desc  string `yaml:",omitempty"`
}

// IsChanged returns true if a Label should be updated.
func (label Label) IsChanged(state Label) bool {
	return label != state
}

// IsEmpty returns true if a Label should be deleted.
func (label Label) IsEmpty() bool {
	return label.Name == "" || label.Color == ""
}

// IsNew returns true if a Label should be created.
func (label Label) IsNew() bool {
	return label.ID == 0
}

// LabelSet represents a GitHub repository set of labels.
type LabelSet struct {
	Name   string
	Labels []Label `yaml:",omitempty"`
	From   []Label `yaml:",omitempty"`
}

func (set LabelSet) Len() int {
	return len(set.Labels)
}

func (set LabelSet) Less(i, j int) bool {
	if set.Labels[i].ID == set.Labels[j].ID {
		return set.Labels[i].Name < set.Labels[j].Name
	}
	return set.Labels[i].ID < set.Labels[j].ID
}

func (set LabelSet) Swap(i, j int) {
	set.Labels[i], set.Labels[j] = set.Labels[j], set.Labels[i]
}

// SortLabelsByName allows to sort labels by name
// instead of ID by default.
type SortLabelsByName LabelSet

func (set SortLabelsByName) Len() int {
	return len(set.Labels)
}

func (set SortLabelsByName) Less(i, j int) bool {
	return set.Labels[i].Name < set.Labels[j].Name
}

func (set SortLabelsByName) Swap(i, j int) {
	set.Labels[i], set.Labels[j] = set.Labels[j], set.Labels[i]
}
