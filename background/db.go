package background

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// NewRepository returns a job repository.
func NewRepository(c Config, db *sqlx.DB) *Repository {
	return &Repository{c, db}
}

// Repository handles DB access to the Job aggregate
type Repository struct {
	config Config
	db     *sqlx.DB
}

// Begin starts a transaction.
func (r *Repository) Begin() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

// AssembleInfrastructure creates the tables for the repository.
func (r *Repository) AssembleInfrastructure() error {
	jobTable := `CREATE TABLE IF NOT EXISTS jobs(
		ID uuid NOT NULL PRIMARY KEY,
		name VARCHAR(256) NOT NULL,
		frequency INT DEFAULT null,
		system_job BOOLEAN NOT NULL DEFAULT false,
		task BYTEA NOT NULL,
		user_id uuid DEFAULT NULL,
		worker uuid DEFAULT NULL,
		heartbeat TIMESTAMP,
		active BOOLEAN NOT NULL DEFAULT true,
		start_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT now()
	)`

	executionTable := `CREATE TABLE IF NOT EXISTS job_executions(
		ID uuid NOT NULL PRIMARY KEY,
		job_id uuid NOT NULL REFERENCES jobs(ID) ON DELETE CASCADE,
		status VARCHAR(10) NOT NULL DEFAULT 'waiting',
		next TIMESTAMP NOT NULL,
		completed_at TIMESTAMP DEFAULT null,
		created_at TIMESTAMP NOT NULL DEFAULT now()
	)`

	_, err := r.db.Exec(jobTable)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(executionTable)
	if err != nil {
		return err
	}

	return nil
}

// Store stores a job in the repository
func (r *Repository) Store(job Job) error {
	query := `
		INSERT INTO jobs(ID, name, frequency, system_job, task,
		user_id, worker, heartbeat, active, start_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT(id)
		DO UPDATE
		SET name = $2, frequency = $3, system_job = $4, task = $5,
		user_id = $6, worker = $7, heartbeat = $8, active = $9,
		start_at = $10`

	tx, err := r.db.Beginx()
	if err != nil {
		panic(err)
	}
	_, err = tx.Exec(query, job.ID, job.Name, job.Frequency, job.SystemJob,
		job.Task, job.UserID, job.Worker, job.Heartbeat, job.Active, job.StartAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	executionInsert := `
	INSERT INTO job_executions(
		ID, job_id, status, next, completed_at
	) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT(id)
	DO UPDATE
	SET status = $3, next = $4, completed_at = $5`

	for _, ex := range job.Executions {
		_, err := tx.Exec(executionInsert,
			ex.ID, ex.JobID, ex.Status, ex.Next, ex.CompletedAt,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetOne retrieves one Job aggregate.
func (r *Repository) GetOne(id uuid.UUID) (Job, error) {
	var job Job
	err := r.db.Get(&job, `SELECT * FROM jobs WHERE ID = $1`, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return job, JobNotFound
	} else if err != nil {
		return job, err
	}

	err = r.db.Select(&job.Executions,
		`SELECT * FROM job_executions WHERE job_id = $1`, id)
	if err != nil {
		return job, err
	}

	for ind := range job.Executions {
		job.Executions[ind].Job = job
	}

	return job, nil
}

// All returns all Job aggregates.
func (r *Repository) All() ([]Job, error) {
	var jobs []Job

	err := r.db.Select(&jobs, `SELECT * FROM jobs`)
	if err != nil {
		return jobs, err
	}

	jobs, err = r.addExecutions(jobs)
	return jobs, err
}

// addExecutions adds executions to a slice of jobs
func (r *Repository) addExecutions(jobs []Job) ([]Job, error) {
	if len(jobs) == 0 {
		return jobs, nil
	}

	jobIds := make([]uuid.UUID, len(jobs))
	jobMap := make(map[uuid.UUID]*Job)
	for ind, job := range jobs {
		jobIds[ind] = job.ID
		jobMap[job.ID] = &jobs[ind]
	}

	var executions []JobExecution
	query, args, err := sqlx.In(`
		SELECT * FROM job_executions
			WHERE job_id IN (?)`, jobIds)
	if err != nil {
		return jobs, err
	}
	query = r.db.Rebind(query)

	err = r.db.Select(&executions, query, args...)
	if err != nil {
		return jobs, err
	}

	for _, execution := range executions {
		execution.Job = *jobMap[execution.JobID]
		jobMap[execution.JobID].Executions = append(jobMap[execution.JobID].Executions, execution)
	}
	return jobs, nil
}

// ClaimFor claims any unclaimed/abandoned jobs for a worker
func (r *Repository) ClaimFor(workerID uuid.UUID) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	var jobIDs []uuid.UUID
	err = tx.Select(&jobIDs, `SELECT ID from jobs 
		WHERE worker is null OR heartbeat < now() FOR UPDATE`)
	if len(jobIDs) == 0 {
		tx.Rollback()
		return nil
	}

	query := `UPDATE jobs
		SET worker = $1, heartbeat = $2
		WHERE worker is null OR heartbeat < now()`

	_, err = tx.Exec(query, workerID, time.Now().Add(time.Minute*10))
	if err != nil {
		return err
	}

	query, jobs, err := sqlx.In(`UPDATE job_executions
		SET status = 'waiting' WHERE status = 'processing' and job_id IN (?)`, jobIDs)
	if err != nil {
		tx.Rollback()
		return err
	}
	query = tx.Rebind(query)

	_, err = tx.Exec(query, jobs...)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}

// Heartbeat updates the heartbeat for all of a worker's jobs.
func (r *Repository) Heartbeat(workerID uuid.UUID) error {
	query := `UPDATE jobs
		SET heartbeat = $1
		WHERE worker = $2`

	nextHeartbeat := time.Now().Add(time.Minute * 10)
	_, err := r.db.Exec(query, nextHeartbeat, workerID)

	return err
}

// GetFor retrieves a worker's jobs.
func (r *Repository) GetFor(workerID uuid.UUID) ([]Job, error) {
	var jobs []Job

	err := r.db.Select(&jobs, `SELECT * FROM jobs WHERE worker = $1`, workerID)
	if err != nil {
		return jobs, err
	}

	jobs, err = r.addExecutions(jobs)
	return jobs, err
}

// StoreAll stores a slice of Job aggregates.
func (r *Repository) StoreAll(jobs []Job) error {
	for _, job := range jobs {
		err := r.Store(job)
		if err != nil {
			return err
		}
	}
	return nil
}
