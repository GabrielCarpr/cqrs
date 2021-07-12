// +build !unit

package background

import (
	"context"
	"encoding/gob"
	"github.com/GabrielCarpr/cqrs/bus"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupCtrl(t *testing.T) (*Repository, *Controller) {
	_, repo := setup(t)
	ctrl := NewController(repo)

	return repo, ctrl
}

func TestCtrlRunsEmptyWithoutPanic(t *testing.T) {
	_, ctrl := setupCtrl(t)
	ctrl.LoopSeconds = 1

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	ctrl.Run(ctx)
}

func TestCtrlHappyPath(t *testing.T) {
	testVal := ""

	repo, ctrl := setupCtrl(t)
	ctrl.LoopSeconds = 1

	cmd := TestCmd{Return: "Hi"}
	gob.Register(cmd)
	job := NewJob("test", cmd)
	job.StartAt = time.Now()
	job.Frequency = 0
	err := repo.Store(job)
	assert.Nil(t, err)

	ctrl.RegisterQueueAction(func(ctx context.Context, c bus.Command) error {
		testVal = "hello world"
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	go func() {
		ctrl.Run(ctx)
	}()
	time.Sleep(time.Second * 2)
	ctrl.FinishTaskForJob(job.ID)
	time.Sleep(time.Second)
	cancel()

	job, err = repo.GetOne(job.ID)
	assert.Nil(t, err)

	c, err := job.decode()
	assert.NoError(t, err)
	tc := c.(TestCmd)
	assert.Equal(t, "Hi", tc.Return)
	assert.Equal(t, COMPLETE, job.NextExecutionStatus())
	assert.WithinDuration(t, time.Now(), job.Executions[0].CompletedAt, time.Second*10)
	assert.Equal(t, "hello world", testVal)
}
