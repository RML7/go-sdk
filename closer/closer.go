package closer

import (
	"context"
	"fmt"
	"log/slog"
)

var globalCloser = closer{
	funcs: make([]func(ctx context.Context) error, 0),
}

func AddFunc(f ...func(ctx context.Context) error) {
	globalCloser.addFunc(f...)
}

func CloseAll(ctx context.Context) {
	globalCloser.closeAll(ctx)
}

type closer struct {
	funcs []func(ctx context.Context) error
}

func (c *closer) addFunc(f ...func(ctx context.Context) error) {
	c.funcs = append(c.funcs, f...)
}

func (c *closer) closeAll(ctx context.Context) {
	for i := len(c.funcs) - 1; i >= 0; i-- {
		if err := c.funcs[i](ctx); err != nil {
			slog.Error(fmt.Sprintf("error close: %s", err.Error()))
		}
	}
}
