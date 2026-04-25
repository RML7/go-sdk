package xslog

import (
	"context"
	"log/slog"
)

type Config struct {
	// ContextAttrs maps log attribute names to context keys.
	ContextAttrs map[string]interface{}
}

type ContextHandler struct {
	slog.Handler
	cfg Config
}

func NewContextHandler(base slog.Handler, cfg Config) *ContextHandler {
	return &ContextHandler{
		Handler: base,
		cfg:     cfg,
	}
}

func (h *ContextHandler) Handle(ctx context.Context, record slog.Record) error {
	for attrName, ctxKey := range h.cfg.ContextAttrs {
		val := ctx.Value(ctxKey)
		if val == nil {
			continue
		}
		record.AddAttrs(slog.Any(attrName, val))
	}

	return h.Handler.Handle(ctx, record)
}
