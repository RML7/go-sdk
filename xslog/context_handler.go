package xslog

import (
	"context"
	"log/slog"
)

type contextHandler struct {
	slog.Handler

	ContextAttrs map[string]interface{}
}

func newContextHandler(base slog.Handler, ContextAttrs map[string]interface{}) *contextHandler {
	return &contextHandler{
		Handler:      base,
		ContextAttrs: ContextAttrs,
	}
}

func (h *contextHandler) Handle(ctx context.Context, record slog.Record) error {
	for attrName, ctxKey := range h.ContextAttrs {
		val := ctx.Value(ctxKey)
		if val == nil {
			continue
		}
		record.AddAttrs(slog.Any(attrName, val))
	}

	return h.Handler.Handle(ctx, record)
}
