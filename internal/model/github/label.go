package github

// Label represents a GitHub label.
type Label struct {
	ID    int64  `yaml:",omitempty"`
	Name  string `yaml:",omitempty"`
	Color string `yaml:",omitempty"`
	Desc  string `yaml:",omitempty"`
}

// Apply modifies the Label with data from the diff.
func (label *Label) Apply(diff LabelPatch) {
	label.Name = diff.Name
	label.Color = diff.Color
	label.Desc = diff.Desc
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

// LabelPatch represents a preset Label.
type LabelPatch struct {
	Label `yaml:",inline"`
	From  []Label `yaml:",omitempty"`
}

// Match calculates score from comparing with Label.
func (patch LabelPatch) Match(with Label) int {
	compare := func(left, right Label) int {
		var score int
		if left.Name == right.Name {
			score += 5
		}
		if left.Color == right.Color {
			score += 1
		}
		if left.Desc == right.Desc {
			score += 1
		}
		return score
	}

	score := compare(patch.Label, with)
	for _, label := range patch.From {
		local := compare(label, with)
		if local > score {
			score = local
		}
	}

	return score
}

// LabelPreset represents a specific set of labels.
type LabelPreset struct {
	Name   string
	Labels []LabelPatch `yaml:",omitempty"`
}

// ExtractMatched founds the most appropriate LabelPatch
// for the specified Label.
func (set *LabelPreset) ExtractMatched(target Label) LabelPatch {
	idx, max := -1, 0

	for i, label := range set.Labels {
		score := label.Match(target)
		if max < score {
			idx, max = i, score
		}
	}

	// no have replacement, delete it
	if idx == -1 {
		return LabelPatch{}
	}
	// have one, update
	diff := set.Labels[idx]
	set.Labels = append(set.Labels[:idx], set.Labels[idx+1:]...)
	return diff
}

// LabelSet represents a GitHub repository set of labels.
type LabelSet struct {
	Name   string
	Labels []Label `yaml:",omitempty"`
}

// FindByID finds a Label by id or returns nil.
func (set LabelSet) FindByID(id int64) *Label {
	for _, label := range set.Labels {
		if label.ID == id {
			label := label
			return &label
		}
	}
	return nil
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
