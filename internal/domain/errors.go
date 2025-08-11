package domain

import "errors"

var (
	ErrBookNotFound      = errors.New("book not found")
	ErrRepositoryError   = errors.New("repository error")
	ErrBookAlreadyExists = errors.New("book already exists")
)
