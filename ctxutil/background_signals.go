package ctxutil

import (
	"context"
	"os"
	"os/signal"
)

// BackgroundWithSignals returns a Context that will be
// canceled with the process receives a SIGINT signal.
// This function starts a goroutine and listens for signals.
func BackgroundWithSignals() context.Context {
	ctx, cancelFn := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		signal.Reset(os.Interrupt)
		cancelFn()
	}()
	return ctx
}
