package ports

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GabrielCarpr/cqrs/log"
	"golang.org/x/sync/errgroup"
)

// Port is an external input to the system that listens, blocking.
//
// The port interface allows an app to concurrently run multiple blocking
// ports while handling cancellation and graceful shutdown.
// The port interface also requires that:
// - The port will only return an error if it cannot continue. An error will force the whole system to shut down
// - The port must block
// - The port will gracefully stop upon the context cancelling
type Port interface {
	Run(context.Context) error
}

// Ports is a collection of entry ports into the system
type Ports []Port

// Run runs all the ports with graceful shutdown.
//
// Run will block, running all the ports concurrently, until receiving an ctx cancellation,
// an OS cancellation signal, or a port returns an error (see Port). Then, it will cancel all other ports.
// If a port fails to exit after 10 seconds, it will forcibly exit. The program should exit shortly after otherwise
// a goroutine leak is possible.
//
// If a port exited with an error, this will be returned, un
func (p Ports) Run(ctx context.Context) error {
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)
	g, ctx := errgroup.WithContext(ctx)

	for _, port := range p {
		g.Go(func() error {
			return port.Run(ctx)
		})
	}

	<-ctx.Done()
	log.Info(ctx, "Quitting - waiting for all ports to exit", log.F{})

	var err error
	ended := make(chan struct{}, 1)
	go func() {
		err = g.Wait()
		ended <- struct{}{}
	}()

	select {
	case <-ended:
		return err
	case <-time.After(time.Second * 10):
		err = fmt.Errorf("Failed to quit after 10 seconds, forced: %w", err)
		break
	}
	return err
}
