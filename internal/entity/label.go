package entity

// Label represents a GitHub label.
type Label struct {
	ID    int64
	Name  string
	Color string
	Desc  string
}

type LabelOperation byte

const (
	SkipLabel LabelOperation = 1 << iota
	CreateLabel
	DeleteLabel
	UpdateLabel
)

func (op LabelOperation) Skip() bool {
	return op == SkipLabel
}

func (op LabelOperation) Create() bool {
	return op == CreateLabel
}

func (op LabelOperation) Delete() bool {
	return op == DeleteLabel
}

func (op LabelOperation) Update() bool {
	return op == UpdateLabel
}

type LabelTransform struct {
	LabelOperation
	From Label
	To   Label
}
