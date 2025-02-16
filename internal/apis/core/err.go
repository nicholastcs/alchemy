package core

import "fmt"

type pathError struct {
	path string
	err  error
}

func (e pathError) Error() string {
	return fmt.Sprintf("at %s: %s", e.path, e.err.Error())
}

func NewPathError(path string, err error) error {
	if path == "" {
		panic("path cannot be empty")
	}

	if err == nil {
		panic("err cannot be nil")
	}

	return &pathError{
		path: path,
		err:  err,
	}
}
