package closer

import (
	"context"
	"fmt"
	"log/slog"
)

var globalCloser = closer{
	funcsWithContext: make([]func(ctx context.Context) error, 0),
	funcs:            make([]func() error, 0),
}

func AddFunc(f ...func() error) {
	globalCloser.addFunc(f...)
}

func AddFuncWithContext(f ...func(ctx context.Context) error) {
	globalCloser.addFuncWithContext(f...)
}

func CloseAll(ctx context.Context) {
	globalCloser.closeAll(ctx)
}

type closer struct {
	funcs            []func() error
	funcsWithContext []func(ctx context.Context) error
}

func (c *closer) addFuncWithContext(f ...func(ctx context.Context) error) {
	c.funcsWithContext = append(c.funcsWithContext, f...)
}

func (c *closer) addFunc(f ...func() error) {
	c.funcs = append(c.funcs, f...)
}

func (c *closer) closeAll(ctx context.Context) {
	for _, f := range c.funcs {
		if err := f(); err != nil {
			slog.Error(fmt.Sprintf("error close: %s", err.Error()))
		}
	}

	for _, f := range c.funcsWithContext {
		if err := f(ctx); err != nil {
			slog.Error(fmt.Sprintf("error close: %s", err.Error()))
		}
	}
}
