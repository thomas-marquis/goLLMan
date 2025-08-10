package domain

type Status string

const (
	StatusUnknown  Status = ""
	StatusError    Status = "error"
	StatusIndexed  Status = "indexed"
	StatusIndexing Status = "indexing"
)

type Book struct {
	ID       string
	Title    string
	Author   string
	Metadata map[string]any
	Selected bool
	Status   Status
}
