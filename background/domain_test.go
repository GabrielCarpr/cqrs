package background

import (
	"cqrs/bus"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type TestCmd struct {
	bus.CommandType

	Return string
}

func (c TestCmd) Valid() error {
	return nil
}

func (c TestCmd) Command() string {
	return "testcmd"
}

func TestNextExecutionStatusUnscheduled(t *testing.T) {
	job := NewJob("test task", TestCmd{})

	assert.Equal(t, job.NextExecutionStatus(), NONE, "Status should be none")
}

func TestNextExecutionStatusScheduled(t *testing.T) {
	job := NewJob("Test task", TestCmd{})
	job.StartAt = time.Now().Add(time.Minute * 5)
	err := job.ScheduleNextExecution()
	assert.Nil(t, err, "Schedule should not error")

	assert.Equal(t, WAITING, job.NextExecutionStatus(), "Status should be waiting")
}

func TestScheduleNextExecution(t *testing.T) {
	tests := []struct {
		name               string
		active             bool
		start              time.Time
		frequency          int
		executions         int
		err                error
		expectedExecutions int
		next               time.Time
	}{
		{
			"Inactive job",
			false,
			time.Time{},
			0,
			0,
			JobNotActive,
			0,
			time.Time{},
		},
		{
			"No start time",
			true,
			time.Time{},
			20,
			0,
			JobNoStartTime,
			0,
			time.Time{},
		},
		{
			"One shot complete",
			true,
			time.Now().Add(-time.Minute * 2),
			0,
			1,
			OneShotJobUsed,
			1,
			time.Time{},
		},
		{
			"First run one shot job",
			true,
			time.Now().Add(time.Minute * 2),
			0,
			0,
			nil,
			1,
			time.Now().Add(time.Minute * 2),
		},
		{
			"Recurring job first run",
			true,
			time.Now(),
			20,
			0,
			nil,
			1,
			time.Now(),
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			job := NewJob(c.name, TestCmd{})
			job.Active = c.active
			job.StartAt = c.start
			job.Frequency = c.frequency

			for i := 0; i < c.executions; i++ {
				job.Executions = append(job.Executions, JobExecution{
					ID:          uuid.New(),
					JobID:       job.ID,
					Status:      COMPLETE,
					Job:         job,
					CompletedAt: time.Now(),
				})
			}

			err := job.ScheduleNextExecution()
			assert.Equal(t, c.err, err, "Error return wrong")
			assert.Equal(t, c.expectedExecutions, len(job.Executions))
			if !c.next.IsZero() {
				assert.WithinDuration(t, c.next, job.NextExecution().Next, time.Second*10)
			}
		})
	}
}

func TestRecurringJobNextRun(t *testing.T) {
	job := NewJob("test", TestCmd{})
	job.StartAt = time.Now().Add(-time.Hour * 1)
	job.Frequency = 70
	job.Executions = []JobExecution{{
		ID:          uuid.New(),
		JobID:       job.ID,
		Status:      COMPLETE,
		Next:        time.Now().Add(-time.Hour),
		Job:         job,
		CreatedAt:   time.Now().Add(-time.Hour),
		ScheduledAt: time.Now().Add(-time.Hour),
		CompletedAt: time.Now().Add(-time.Hour),
	}}

	job.ScheduleNextExecution()

	assert.Equal(t, WAITING, job.NextExecutionStatus())

	next := job.NextExecution()

	assert.WithinDuration(t, time.Now().Add(time.Minute*10).Round(time.Minute*70), next.Next, time.Minute)
}

func TestMissedRecurringJob(t *testing.T) {
	job := NewJob("test", TestCmd{})
	job.StartAt = time.Now().Add(-time.Hour * 3)
	job.Frequency = 60
	job.Executions = []JobExecution{{
		ID:          uuid.New(),
		JobID:       job.ID,
		Status:      COMPLETE,
		Next:        time.Now().Add(-time.Hour * 2),
		Job:         job,
		CreatedAt:   time.Now().Add(-time.Hour * 2),
		ScheduledAt: time.Now().Add(-time.Hour * 2),
		CompletedAt: time.Now().Add(-time.Hour * 2),
	}}

	job.ScheduleNextExecution()

	assert.Equal(t, WAITING, job.NextExecutionStatus())
	next := job.NextExecution()
	assert.WithinDuration(t, time.Now().Round(time.Hour), next.Next, time.Minute)
}

func TestScheduleNextExecutionOneWaiting(t *testing.T) {
	job := NewJob("test", TestCmd{})
	job.StartAt = time.Now().Add(-time.Minute * 5)
	job.Frequency = 10
	assert.Nil(t, job.ScheduleNextExecution())
	assert.Error(t, ExecutionAlreadyWaiting, job.ScheduleNextExecution())
}

func TestIsDue(t *testing.T) {
	tests := []struct {
		name      string
		status    ExecutionStatus
		startAt   time.Time
		frequency int
		expected  bool
	}{
		{
			"Not waiting",
			NONE,
			time.Now().Add(time.Minute * 2),
			0,
			false,
		},
		{
			"First execution one shot",
			WAITING,
			time.Now().Add(-time.Minute * 2),
			0,
			true,
		},
		{
			"First execution recurring",
			WAITING,
			time.Now().Add(-time.Minute * 5),
			20,
			true,
		},
		{
			"Already scheduled",
			PROCESSING,
			time.Now().Add(-time.Minute - 5),
			20,
			false,
		},
		{
			"One shot complete",
			COMPLETE,
			time.Now().Add(-time.Hour * 2),
			0,
			false,
		},
		{
			"Recurring future",
			WAITING,
			time.Now().Add(time.Hour * 3),
			500,
			false,
		},
		{
			"One shot future",
			WAITING,
			time.Now().Add(time.Hour * 5),
			0,
			false,
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			job := NewJob("test", TestCmd{})
			job.Frequency = c.frequency
			job.StartAt = c.startAt

			if c.status != NONE {
				job.addJobExecution(c.startAt)
				job.Executions[0].Status = c.status
			}

			assert.Equal(t, c.expected, job.IsDue())
		})
	}
}

func TestIsRecurring(t *testing.T) {
	job := NewJob("test", TestCmd{})
	job.Frequency = 5
	assert.True(t, job.isRecurring())

	job.Frequency = 0
	assert.False(t, job.isRecurring())
}

func TestCompleteOneTime(t *testing.T) {
	workerID := uuid.New()
	job := NewJob("test", TestCmd{})
	job.StartAt = time.Now()
	job.Frequency = 0
	job.Worker = workerID
	job.ScheduleNextExecution()

	err := job.Complete(workerID)
	assert.Nil(t, err)

	assert.False(t, job.Active)
	next := job.NextExecution()
	assert.Equal(t, COMPLETE, next.Status)
	assert.WithinDuration(t, time.Now(), next.CompletedAt, time.Second*2)
}

func TestCompleteRecurring(t *testing.T) {
	workerID := uuid.New()
	job := NewJob("test", TestCmd{})
	job.StartAt = time.Now()
	job.Frequency = 20
	job.Worker = workerID
	job.ScheduleNextExecution()

	err := job.Complete(workerID)
	assert.Nil(t, err)

	assert.True(t, job.Active)
	next := job.NextExecution()
	assert.Equal(t, WAITING, next.Status)
	assert.WithinDuration(t, time.Now().Add(time.Minute*20).Round(time.Minute*20), next.Next, time.Second*5)
	assert.Equal(t, 2, len(job.Executions))
	assert.WithinDuration(t, time.Now(), job.Executions[0].CompletedAt, time.Second*2)
}

func TestScheduleNowErrors(t *testing.T) {
	job := NewJob("test", TestCmd{})
	job.StartAt = time.Now().Add(time.Minute * 5)
	job.ScheduleNextExecution()

	assert.Error(t, JobNotDue, job.ScheduleNow())
}

func TestScheduleNow(t *testing.T) {
	job := NewJob("test", TestCmd{})
	job.StartAt = time.Now()
	job.ScheduleNextExecution()

	assert.Nil(t, job.ScheduleNow())

	next := job.NextExecution()

	assert.Equal(t, PROCESSING, next.Status)
	assert.WithinDuration(t, time.Now(), next.ScheduledAt, time.Second)
}
