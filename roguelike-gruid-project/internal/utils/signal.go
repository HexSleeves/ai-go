package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"codeberg.org/anaseto/gruid"
)

// HandleSignals sets up signal handling for graceful shutdown
func HandleSignals(ctx context.Context, msgs chan<- gruid.Msg) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)
	select {
	case <-ctx.Done():
	case <-sig:
		msgs <- gruid.MsgQuit{}
	}
}
