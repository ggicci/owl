package in

import (
	"context"
	"errors"
)

type Core struct {
	ctx context.Context
}

func New(inputStruct interface{}) (*Core, error) {
	return nil, errors.New("not implemented")
}

func (c *Core) Context() context.Context {
	return c.ctx
}

func (c *Core) Decode()
