package domain

type File struct {
	Name string
}

type FileWithContent struct {
	File
	Content []byte
}
