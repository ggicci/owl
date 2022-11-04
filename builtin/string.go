package builtin

import "errors"

type String struct {
	base
	Value string
}

func (s *String) ParseFrom(values []string) error {
	return errors.New("not implemented")
}
