package builtin

import "errors"

type Int struct {
	base
	Value int
}

func (i *Int) ParseFrom(values []string) error {
	return errors.New("not implemented")
}
