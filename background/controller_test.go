// +build !unit

package background

import (
	"context"
	"cqrs/bus"
	"encoding/gob"
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

	ctrl.Block(2)
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
	ctrl.Block(2)
	ctrl.FinishTaskForJob(job.ID)
	ctrl.Block(1)

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
