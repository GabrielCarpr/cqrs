package background

import (
	"bytes"
	"cqrs/bus"
	"encoding/gob"
	"time"

	"github.com/google/uuid"
)

const (
	WAITING    ExecutionStatus = "waiting"
	PROCESSING ExecutionStatus = "processing"
	COMPLETE   ExecutionStatus = "complete"
	NONE       ExecutionStatus = "none"
)

// ExecutionStatus is the current status of one Job Execution.
type ExecutionStatus string

/**
 * Job
 */

// NewJob creates a job with the basic legal defaults.
func NewJob(name string, cmd bus.Command) Job {
	payload, _ := encode(cmd)
	return Job{
		ID:        uuid.New(),
		Name:      name,
		Task:      payload,
		Active:    true,
		SystemJob: true,
	}
}

// Job is a domain entity for a Job, delayed execution task.
type Job struct {
	ID        uuid.UUID
	Name      string
	Frequency int
	SystemJob bool `json:"system_job" db:"system_job"`
	Task      []byte
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Worker    uuid.UUID
	Heartbeat time.Time
	Active    bool
	StartAt   time.Time `json:"start_at" db:"start_at"`

	Executions []JobExecution `db:"-"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// isRecurring returns if the job is a recurring job, otherwise it's a one time job.
func (j Job) isRecurring() bool {
	return j.Frequency > 0
}

// Complete updates the job after it has finished
func (j *Job) Complete(workerID uuid.UUID) error {
	j.Executions[len(j.Executions)-1].Status = COMPLETE
	j.Executions[len(j.Executions)-1].CompletedAt = time.Now()

	if !j.isRecurring() {
		j.Active = false
	} else {
		err := j.ScheduleNextExecution()
		if err != nil {
			return err
		}
	}
	return nil
}

// NextExecutionStatus returns the status of the next
// (aka currently pending) execution, or NONE
func (j Job) NextExecutionStatus() ExecutionStatus {
	last := j.NextExecution()
	if last == nil {
		return NONE
	}
	return last.Status
}

// NextExecution returns the next execution
func (j Job) NextExecution() *JobExecution {
	if len(j.Executions) == 0 {
		return nil
	}
	return &j.Executions[len(j.Executions)-1]
}

// IsDue returns whether the job is due to be scheduled (aka queued)
func (j Job) IsDue() bool {
	return (j.NextExecutionStatus() == WAITING &&
		j.Active &&
		j.NextExecution().Next.Before(time.Now()))
}

// ScheduleNow modifies the job after it's been scheduled/queued
func (j *Job) ScheduleNow() error {
	if !j.IsDue() {
		return JobNotDue
	}

	exeIndex := len(j.Executions) - 1

	j.Executions[exeIndex].Status = PROCESSING
	j.Executions[exeIndex].ScheduledAt = time.Now()
	return nil
}

// ScheduleNextExecution creates a the next execution
func (j *Job) ScheduleNextExecution() error {
	if j.Active != true {
		return JobNotActive
	}
	if j.StartAt.IsZero() {
		return JobNoStartTime
	}
	if j.Frequency == 0 && len(j.Executions) > 0 {
		return OneShotJobUsed
	}

	var next time.Time
	if j.Frequency == 0 {
		next = j.StartAt
	} else if j.NextExecutionStatus() == NONE {
		next = j.StartAt
	} else if j.NextExecutionStatus() == COMPLETE {
		next = j.calculateNextIteration()
	} else {
		return ExecutionAlreadyWaiting
	}

	j.addJobExecution(next)
	return nil
}

func (j Job) calculateNextIteration() time.Time {
	freqDuration := time.Minute * time.Duration(j.Frequency)
	lastScheduled := j.NextExecution().Next
	proposed := lastScheduled.Add(freqDuration)

	if proposed.Before(time.Now()) {
		return time.Now().Round(freqDuration)
	}
	return proposed.Round(freqDuration)
}

func encode(c bus.Command) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(&c)
	return buf.Bytes(), err
}

func (j Job) decode() (bus.Command, error) {
	buf := bytes.NewBuffer(j.Task)
	var c bus.Command
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&c)
	return c, err
}

/**
 * Job Execution
 */

func (j *Job) addJobExecution(next time.Time) {
	execution := JobExecution{
		ID:     uuid.New(),
		JobID:  j.ID,
		Status: WAITING,
		Next:   next,
		Job:    *j,
	}
	j.Executions = append(j.Executions, execution)
}

// JobExecution is a domain entity for one run of a Job.
type JobExecution struct {
	ID     uuid.UUID
	JobID  uuid.UUID `json:"job_id" db:"job_id"`
	Status ExecutionStatus
	Next   time.Time

	Job Job `db:"-"`

	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	ScheduledAt time.Time `json:"scheduled_at" db:"scheduled_at"`
	CompletedAt time.Time `json:"completed_at" db:"completed_at"`
}
