package domain

type Book struct {
	ID       string
	Title    string
	Author   string
	Metadata map[string]any
}
