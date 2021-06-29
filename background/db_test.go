// +build !unit

package background

import (
	"encoding/gob"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

var db *sqlx.DB

func testConfig() Config {
	return Config{
		AppName: "testing",
		DBHost:  "db",
		DBUser:  "cqrs",
		DBPass:  "cqrs",
		DBName:  "cqrs",
	}
}

func getDB() *sqlx.DB {
	config := testConfig()
	db := sqlx.MustConnect("postgres",
		fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=disable",
			config.DBUser, config.DBPass, config.DBName, config.DBHost))
	return db
}

func setup(t *testing.T) (*sqlx.DB, *Repository) {
	config := testConfig()

	if db == nil {
		db = getDB()
	}
	_, err := db.Exec("DROP TABLE IF EXISTS job_executions; DROP TABLE IF EXISTS jobs;")
	if err != nil {
		t.Errorf("Failed dropping: %s", err)
	}
	repo := NewRepository(config, db)
	err = repo.AssembleInfrastructure()
	if err != nil {
		t.Errorf("Failed migrating: %s", err)
	}

	gob.Register(TestCmd{})

	return db, repo
}

func failOnErr(t *testing.T, message string, err error) {
	if err != nil {
		t.Errorf("%s: %s", message, err)
	}
}

func TestStoreAndGetOne(t *testing.T) {
	_, repo := setup(t)

	job := NewJob("test task", TestCmd{})
	job.StartAt = time.Now().Add(time.Second * 1)
	failOnErr(t, "Failed scheduling", job.ScheduleNextExecution())

	err := repo.Store(job)
	failOnErr(t, "Failed storing", err)

	newJob, err := repo.GetOne(job.ID)
	failOnErr(t, "Failed finding", err)

	if newJob.Name != "test task" {
		t.Errorf("Wrong name: %s", newJob.Name)
	}
	if len(newJob.Executions) != 1 {
		t.Errorf("Wrong job executions: %v", len(newJob.Executions))
	}
}

func TestGetNone(t *testing.T) {
	_, repo := setup(t)

	_, err := repo.GetOne(uuid.New())

	assert.Error(t, JobNotFound, err)
}

func TestGetAll(t *testing.T) {
	_, repo := setup(t)

	job1 := NewJob("test 1", TestCmd{})
	job1.StartAt = time.Now().Add(time.Minute * 5)
	failOnErr(t, "Failed scheduling", job1.ScheduleNextExecution())
	if len(job1.Executions) != 1 {
		t.Errorf("Execution was not added to job: %v", len(job1.Executions))
	}
	job2 := NewJob("test 2", TestCmd{})
	job3 := NewJob("test 3", TestCmd{})

	failOnErr(t, "Failed storing 1", repo.Store(job1))
	failOnErr(t, "Failed storing 2", repo.Store(job2))
	failOnErr(t, "Failed storing 3", repo.Store(job3))

	jobs, err := repo.All()
	failOnErr(t, "Failed getting all", err)

	if len(jobs) != 3 {
		t.Errorf("Wrong number of jobs: %v", len(jobs))
	}
	if len(jobs[0].Executions) != 1 {
		t.Errorf("Wrong number of job 1 executions: %v %v %v", len(jobs[0].Executions), len(jobs[1].Executions), len(jobs[2].Executions))
	}
}

func TestClaimUnclaimed(t *testing.T) {
	workerID := uuid.New()

	_, repo := setup(t)

	job1 := NewJob("test1", TestCmd{})
	job1.Worker = uuid.New()
	job1.Heartbeat = time.Now().Add(time.Minute * 9)
	job2 := NewJob("test2", TestCmd{})

	failOnErr(t, "Failed storing 1", repo.Store(job1))
	failOnErr(t, "Failed storing 2", repo.Store(job2))

	failOnErr(t, "Failed claiming", repo.ClaimFor(workerID))

	newJob1, err := repo.GetOne(job1.ID)
	failOnErr(t, "Failed getting 1", err)
	newJob2, err := repo.GetOne(job2.ID)
	failOnErr(t, "Failed getting 2", err)

	if newJob1.Worker != job1.Worker {
		t.Errorf("Job 1 stolen")
	}
	if newJob2.Worker != workerID {
		t.Errorf("Job 2 not claimed")
	}
}

func TestClaimAbandoned(t *testing.T) {
	workerID := uuid.New()

	_, repo := setup(t)

	job1 := NewJob("test1", TestCmd{})
	job1.Worker = uuid.New()
	job1.Heartbeat = time.Now().Add(time.Minute * 8)
	job2 := NewJob("test2", TestCmd{})
	job2.Worker = uuid.New()
	job2.Heartbeat = time.Now().Add(-time.Minute * 2)

	failOnErr(t, "Failed storing 1", repo.Store(job1))
	failOnErr(t, "Failed storing 2", repo.Store(job2))
	failOnErr(t, "Failed claiming", repo.ClaimFor(workerID))

	newJob1, err := repo.GetOne(job1.ID)
	failOnErr(t, "Failed getting job1", err)
	newJob2, err := repo.GetOne(job2.ID)
	failOnErr(t, "Failed getting job2", err)

	if newJob1.Worker == workerID {
		t.Errorf("Job 1 stolen")
	}
	if newJob2.Worker != workerID {
		t.Errorf("Job 2 not claimed")
	}
}

func TestClaimExecutions(t *testing.T) {
	workerID := uuid.New()

	_, repo := setup(t)

	job := NewJob("test", TestCmd{})
	job.Worker = uuid.New()
	job.Heartbeat = time.Now().Add(-time.Minute * 2)
	job.StartAt = time.Now()
	job.ScheduleNextExecution()
	job.ScheduleNow()

	assert.Equal(t, PROCESSING, job.NextExecutionStatus())
	assert.NotEqual(t, workerID, job.Worker)

	assert.Nil(t, repo.Store(job))

	assert.Nil(t, repo.ClaimFor(workerID))

	job, err := repo.GetOne(job.ID)
	assert.Nil(t, err)

	assert.Equal(t, workerID, job.Worker)
	assert.Equal(t, WAITING, job.NextExecutionStatus())
}

func TestHeartbeat(t *testing.T) {
	workerID := uuid.New()

	_, repo := setup(t)

	job1 := NewJob("t1", TestCmd{})
	job1.Worker = uuid.New()
	job1.Heartbeat = time.Now().Add(time.Minute * 2)
	job2 := NewJob("t2", TestCmd{})
	job2.Worker = workerID
	job2.Heartbeat = time.Now().Add(time.Minute * 5)
	job3 := NewJob("t3", TestCmd{})
	job3.Worker = workerID
	job3.Heartbeat = time.Now().Add(time.Minute * 5)

	failOnErr(t, "Failed storing 1", repo.Store(job1))
	failOnErr(t, "Failed storing 2", repo.Store(job2))
	failOnErr(t, "Failed storing 3", repo.Store(job3))
	failOnErr(t, "Heartbeat failed", repo.Heartbeat(workerID))

	newJob1, err := repo.GetOne(job1.ID)
	failOnErr(t, "Failed getting job1", err)
	newJob2, err := repo.GetOne(job2.ID)
	failOnErr(t, "Failed getting job2", err)
	newJob3, err := repo.GetOne(job3.ID)
	failOnErr(t, "Failed getting job3", err)

	assert.WithinDuration(t, time.Now().Add(time.Minute*2), newJob1.Heartbeat, time.Second*10)
	assert.WithinDuration(t, time.Now().Add(time.Minute*10), newJob2.Heartbeat, time.Second*10)
	assert.WithinDuration(t, time.Now().Add(time.Minute*10), newJob3.Heartbeat, time.Second*10)
}

func TestGetFor(t *testing.T) {
	workerID := uuid.New()

	_, repo := setup(t)

	job1 := NewJob("t1", TestCmd{})
	job1.Worker = uuid.New()
	job2 := NewJob("t2", TestCmd{})
	job2.Worker = workerID
	job3 := NewJob("t3", TestCmd{})
	job3.Worker = workerID

	failOnErr(t, "Failed storing 1", repo.Store(job1))
	failOnErr(t, "Failed storing 2", repo.Store(job2))
	failOnErr(t, "Failed storing 3", repo.Store(job3))

	jobs, err := repo.GetFor(workerID)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(jobs))
	assert.NotEqual(t, "t1", jobs[0].Name)
	assert.NotEqual(t, "t1", jobs[1].Name)
}

func TestStoreAll(t *testing.T) {
	_, repo := setup(t)

	job1 := NewJob("t1", TestCmd{})
	job2 := NewJob("t2", TestCmd{})
	job3 := NewJob("t3", TestCmd{})

	err := repo.StoreAll([]Job{job1, job2, job3})
	assert.Nil(t, err)
	assert.Nil(t, err)

	jobs, err := repo.All()
	assert.Nil(t, err)

	assert.Equal(t, 3, len(jobs))
}
