package ports_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GabrielCarpr/cqrs/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPort struct {
	exec func(context.Context) error
}

func (t testPort) Run(ctx context.Context) error {
	return t.exec(ctx)
}

func TestPortsRunsAndCancels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := testPort{}
	p.exec = func(c context.Context) error {
		select {
		case <-time.After(time.Second * 3):
			return errors.New("did not cancel")
		case <-c.Done():
			return nil
		}
	}

	pts := ports.Ports{p}
	ctx, _ = context.WithTimeout(ctx, time.Millisecond*20)
	err := pts.Run(ctx)
	require.NoError(t, err)
}

func TestPortCancelsAll(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p1run := false
	p1 := testPort{}
	p1.exec = func(c context.Context) error {
		p1run = true
		select {
		case <-time.After(time.Second * 3):
			return errors.New("did not cancel")
		case <-c.Done():

			return nil
		}
	}

	p2run := false
	p2 := testPort{}
	p2.exec = func(c context.Context) error {
		p2run = true
		select {
		case <-time.After(time.Second * 3):
			return errors.New("did not cancel")
		case <-c.Done():

			return nil
		}
	}

	pts := ports.Ports{p1, p2}
	ctx, _ = context.WithTimeout(ctx, time.Millisecond*50)
	err := pts.Run(ctx)
	assert.NoError(t, err)
	assert.True(t, p1run)
	assert.True(t, p2run)
}

func TestPortErrorCancelsAll(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p1 := testPort{}
	p1.exec = func(c context.Context) error {
		return errors.New("error")
	}

	exitedGracefully := false
	p2 := testPort{}
	p2.exec = func(c context.Context) error {
		select {
		case <-c.Done():
			exitedGracefully = true
			return nil
		case <-time.After(time.Second * 3):
			return errors.New("did not cancel")
		}
	}

	pts := ports.Ports{p1, p2}
	ctx, _ = context.WithTimeout(ctx, time.Millisecond*1000)
	err := pts.Run(ctx)
	require.Error(t, err)
	assert.EqualError(t, err, "error")
	assert.True(t, exitedGracefully)
}

func TestPortPanicCancelsAll(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p1 := testPort{}
	p1.exec = func(c context.Context) error {
		panic("oops")
		return nil
	}

	exitedGracefully := false
	p2 := testPort{}
	p2.exec = func(c context.Context) error {
		select {
		case <-c.Done():
			exitedGracefully = true
			return nil
		case <-time.After(time.Second * 3):
			return errors.New("did not cancel")
		}
	}

	pts := ports.Ports{p1, p2}
	ctx, _ = context.WithTimeout(ctx, time.Millisecond*1000)
	err := pts.Run(ctx)
	require.Error(t, err)
	assert.Error(t, err, "panic: oops")
	assert.True(t, exitedGracefully)
}
