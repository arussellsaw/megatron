package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/monzo/slog"
	"github.com/monzo/typhon"

	"github.com/arussellsaw/megatron/domain/library"
	"github.com/arussellsaw/megatron/handler"
)

var (
	mediaPath = "/home/share/tv/Peep Show/"
)

func main() {
	ctx := context.Background()
	slog.SetDefaultLogger(&logFormat{os.Stdout})

	index, err := library.BuildIndex(ctx, mediaPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	handler.DefaultIndex = index

	r := handler.Router()

	s, err := typhon.Listen(r, ":8000")
	if err != nil {
		slog.Error(ctx, "Error serving: %s", err)
		return
	}
	<-s.Done()
}

var _ slog.Logger = &logFormat{}

type logFormat struct {
	w io.Writer
}

func (l *logFormat) Log(events ...slog.Event) {
	for _, e := range events {
		fmt.Println(e.Severity, e.Message)
	}
}

func (l *logFormat) Flush() error {
	return nil
}
