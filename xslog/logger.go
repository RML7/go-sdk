package xslog

import (
	"io"
	"log/slog"
	"os"
)

type Format int

const (
	FormatJSON Format = iota
	FormatText
)

type LoggerConfig struct {
	Level        slog.Level
	Format       Format
	Writer       io.Writer
	ContextAttrs map[string]interface{}
}

func Init(cfg LoggerConfig) {
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}

	opts := &slog.HandlerOptions{Level: cfg.Level}

	var base slog.Handler
	switch cfg.Format {
	case FormatText:
		base = slog.NewTextHandler(cfg.Writer, opts)
	default:
		base = slog.NewJSONHandler(cfg.Writer, opts)
	}

	slog.SetDefault(slog.New(newContextHandler(base, cfg.ContextAttrs)))
}
