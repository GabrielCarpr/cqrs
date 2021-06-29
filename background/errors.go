package background

import (
	"errors"
	"fmt"
)

var (
	JobConditionalErr       = errors.New("")
	JobNotFound             = fmt.Errorf("Job not found%w", JobConditionalErr)
	JobNotActive            = fmt.Errorf("Job is not active%w", JobConditionalErr)
	JobNoStartTime          = fmt.Errorf("Job has no start time%w", JobConditionalErr)
	OneShotJobUsed          = fmt.Errorf("Job is one-shot and has already ran or been scheduled%w", JobConditionalErr)
	ExecutionAlreadyWaiting = fmt.Errorf("Job already has an execution waiting%w", JobConditionalErr)

	JobNotDue  = errors.New("Job not due")
	JobNotMine = errors.New("Job does not belong to this worker")

	JobTaskUnknown = func(taskName string) error {
		return errors.New(fmt.Sprintf("Job's task is unknown: %s", taskName))
	}

	UnknownJobStatus = func(status ExecutionStatus) error {
		return errors.New(fmt.Sprintf("Job has unknown status: %s", status))
	}
)
