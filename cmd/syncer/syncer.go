package syncer

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"github.com/Bortnyak/file-syncer/pkg/client"
	"github.com/Bortnyak/file-syncer/pkg/config"
	"github.com/Bortnyak/file-syncer/pkg/server"
	"github.com/Bortnyak/file-syncer/pkg/watcher"
	"golang.org/x/sync/errgroup"
)

var ErrTerm = errors.New("termination")

func Run() error {
	log.Println("Syncer is starting.....")

	err := config.LoadConfig()
	if err != nil {
		log.Println("Failed to load config, ", err)
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return watcher.Watch(ctx)
	})

	eg.Go(func() error {
		return server.StartServer(ctx)
	})

	eg.Go(func() error {
		return client.StartClient(ctx)
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
			// Close connections, etc
		} else {
			log.Println("Shutting down due to error: ", err)
		}
	}

	return nil
}
