package syncer

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/Bortnyak/file-syncer/pkg/client"
	"github.com/Bortnyak/file-syncer/pkg/server"
	"github.com/Bortnyak/file-syncer/pkg/watcher"
)

var ErrTerm = errors.New("termination")

func Run() {
	log.Println("starting the app")
	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return watcher.Watch(ctx)
	})
	eg.Go(func() error {
		return server.Main(ctx)
	})
	eg.Go(func() error {
		return client.ListenToUpdates(ctx)
	})
	eg.Go(func() error {
		signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		defer cancel()
		<-signalCtx.Done()

		return ErrTerm
	})

	if err := eg.Wait(); err != nil {
		if errors.Is(err, ErrTerm) {
			log.Println("Gracefully shutting down")
			// TODO: add code for graceful shutdown cleanups here
		} else {
			log.Printf("shutting down due to error: %v\n", err)
		}
	}
}
