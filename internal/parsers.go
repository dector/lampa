package internal

import (
	"errors"
	"strings"
)

type DependenciesTree struct {
}

func ParseTree(source string) (DependenciesTree, error) {
	result := DependenciesTree{}

	source = strings.TrimSpace(source)
	if len(source) == 0 {
		return result, ErrEmptyInput
	}

	return result, nil
}

var ErrEmptyInput = errors.New("input is empty")

func IsEmptyInput(err error) bool {
	return err == ErrEmptyInput
}
