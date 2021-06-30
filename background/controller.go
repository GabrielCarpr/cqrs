package background

import (
	"context"
	"errors"
	"fmt"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/bus/message"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// queueAction is a callback from Messenger that allows JobController to queue tasks.
// TODO: Find some way of interfacing with cqrs cmd type
type queueAction func(context.Context, bus.Command) error

// NewController returns a new controller.
func NewController(r *Repository) *Controller {
	c := &Controller{
		repo: r,
	}
	c.init()
	return c
}

// Controller is the job controller, which runs control loops
type Controller struct {
	LoopSeconds float64

	repo      *Repository
	queueTask queueAction

	workerID uuid.UUID
	actions  []func() error
}

// registerQueueAction attaches the QA callback
func (c *Controller) RegisterQueueAction(qa queueAction) {
	c.queueTask = qa
}

// init configures the controller.
func (c *Controller) init() {
	if c.workerID == uuid.Nil {
		c.workerID = uuid.New()
	}
	log.Printf("Registered job worker ID: %s", c.workerID.String())
	c.LoopSeconds = 5.0
	c.registerActions()
	err := c.repo.AssembleInfrastructure()
	if err != nil {
		panic(err)
	}
}

// Run asynchronously runs the control loop, receiving a done signal.
func (c *Controller) Run(done chan bool) {
	log.Print("Starting job control loop")
	ticker := time.NewTicker(time.Second * time.Duration(c.LoopSeconds))
	failures := 0
	failureLimit := 10

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return

			case <-ticker.C:
				errs := c.runActions()
				for _, err := range errs {
					failures++
					log.Printf("Error during action: %v", err)
				}
				if failures >= failureLimit {
					ticker.Stop()
					panic("Too many errors - stopping")
				}
			}
		}
	}()
}

func (c *Controller) runActions() (errors []error) {
	log.Print("Running scheduled actions")
	defer func() {
		if r := recover(); r != nil {
			errors = []error{fmt.Errorf("Panicked: %v", r)}
		}
	}()
	for _, action := range c.actions {
		err := action()
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (c *Controller) registerActions() {
	c.actions = make([]func() error, 2)

	c.actions[0] = c.manageJobs
	c.actions[1] = c.manageExecutions
}

// Block optionally runs the control loop, while blocking.
func (c *Controller) Block(seconds int) {
	done := make(chan bool)
	c.Run(done)
	if seconds == 0 {
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		done <- true
		return
	}
	time.Sleep(time.Duration(seconds*1000) * time.Millisecond)
	done <- true
	return
}

// manageJobs manages all of the workers job entries
func (c *Controller) manageJobs() error {
	log.Print("Managing jobs")

	err := c.repo.ClaimFor(c.workerID)
	if err != nil {
		return err
	}
	err = c.repo.Heartbeat(c.workerID)
	if err != nil {
		return err
	}

	return nil
}

// manageExecutions manages the worker's job's executions
func (c *Controller) manageExecutions() error {
	log.Print("Managing executions")
	jobs, err := c.repo.GetFor(c.workerID)
	if err != nil {
		return err
	}

	for ind := range jobs {
		err := c.advanceJob(&jobs[ind])
		if err != nil {
			return err
		}
	}
	return nil
}

// advanceJob moves a job to its next status, and handles side effects
func (c *Controller) advanceJob(j *Job) error {
	current := j.NextExecutionStatus()

	var err error
	switch current {
	case NONE:
		err = j.ScheduleNextExecution()
		break

	case WAITING:
		if j.IsDue() {
			j.ScheduleNow()
			// Store before queueing to prevent race condition
			err = c.repo.Store(*j)
			if err != nil {
				return err
			}
			return c.queueJob(*j)
		}
		break

	case PROCESSING:
		return nil // Nothing to do, only to wait for the task to be completed

	case COMPLETE:
		if j.isRecurring() {
			return nil
		}
		err = j.ScheduleNextExecution()
		break

	default:
		return UnknownJobStatus(current)
	}

	err = c.repo.Store(*j)
	if err != nil {
		return err
	}

	if errors.Is(err, JobConditionalErr) {
		return nil
	}
	return err
}

// queueJob queues a due job into the task queue.
func (c *Controller) queueJob(j Job) error {
	payload, err := j.decode()
	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, jobID, j.ID)
	log.Printf("Queueing job: %s", j.Name)
	return c.queueTask(ctx, payload.(bus.Command))
}

// finishTaskForJob records a job as finished.
func (c *Controller) FinishTaskForJob(jobID uuid.UUID) error {
	log.Printf("Marking job completed: %s", jobID)
	job, err := c.repo.GetOne(jobID)
	if err != nil {
		return err
	}

	err = job.Complete(c.workerID)
	if err != nil {
		return err
	}

	return c.repo.Store(job)
}

type backgroundCtxKey string

var jobID backgroundCtxKey = "JobID"

// JobFinishingMiddleware hooks into the bus's command execution
// stack and allows it to report to the controller about the jobs execution status when it passes through.
// Should be inserted ABOVE recovery middleware so that panics don't stop job status being reported
func (c *Controller) JobFinishingMiddleware(next bus.CommandHandler) bus.CommandHandler {
	return bus.CmdMiddlewareFunc(func(ctx context.Context, cmd bus.Command) (res bus.CommandResponse, msgs []message.Message) {
		j := ctx.Value(jobID)
		if j == nil {
			return next.Execute(ctx, cmd)
		}
		jID := j.(uuid.UUID)

		log.Printf("Executing job, ID: %s", jID)
		res, msgs = next.Execute(ctx, cmd)
		if res.Error != nil {
			log.Printf("Tried executing job, failed with error: %v", res.Error)
		} else {
			log.Printf("Finished executing job, ID: %s", jID)
		}

		err := c.FinishTaskForJob(jID)
		if err != nil {
			log.Printf("Tried finishing job, error'd: %s", err)
		}
		return
	})
}
