package domain

type Status string

const (
	StatusUnknown  Status = ""
	StatusNew      Status = "new"
	StatusError    Status = "error"
	StatusIndexed  Status = "indexed"
	StatusIndexing Status = "indexing"
)

func StatusFromString(s string) Status {
	switch s {
	case string(StatusNew):
		return StatusNew
	case string(StatusError):
		return StatusError
	case string(StatusIndexed):
		return StatusIndexed
	case string(StatusIndexing):
		return StatusIndexing
	default:
		return StatusUnknown
	}
}

func (s Status) String() string {
	if s == "" {
		return "unknown"
	}
	return string(s)
}

type Book struct {
	ID       string
	Title    string
	Author   string
	Metadata map[string]any
	Selected bool
	Status   Status
	File     File
}

type BookOption func(b *Book)

func WithStatus(status Status) BookOption {
	return func(b *Book) {
		b.Status = status
	}
}
