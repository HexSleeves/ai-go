package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"codeberg.org/anaseto/gruid"
)

// MsgTerminate is a custom message type for termination signals
type MsgTerminate struct {
	Signal os.Signal
	Time   time.Time
}

// HandleSignals sets up signal handling for graceful shutdown
func HandleSignals(ctx context.Context, msgs chan<- gruid.Msg) {
	slog.Info("Setting up signal handling for graceful shutdown")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)

	slog.Info("Signal handler ready, waiting for signals...")
	select {
	case <-ctx.Done():
		slog.Info("Context done, exiting signal handler")
	case signal := <-sig:
		slog.Info("Signal received in handler", "signal", signal)
		termMsg := MsgTerminate{
			Signal: signal,
			Time:   time.Now(),
		}
		slog.Info("Sending termination message", "message", termMsg)
		msgs <- termMsg
		slog.Info("Termination message sent successfully")
	}
	slog.Info("Signal handler exiting")
}
