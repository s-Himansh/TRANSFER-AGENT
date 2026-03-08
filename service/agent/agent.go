package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"transfer.agent/models"
	service "transfer.agent/service"

	senderSvc "transfer.agent/service/sender"
)

type Agent struct {
	// Configuration
	name       string
	serverAddr string
	workers    int

	// faciltating transfer queue
	jobQueue    chan *models.TransferJob
	statusStore map[string]*models.TransferJob
	statusMutex sync.RWMutex

	// transfer control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// service dependencies
	sender service.Sender
}

// NewAgent initialises memory to a new transfer agent within service
func NewAgent(name, serverAddr string, workers int) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		name:        name,
		serverAddr:  serverAddr,
		workers:     workers,
		jobQueue:    make(chan *models.TransferJob, 100),
		statusStore: make(map[string]*models.TransferJob),
		ctx:         ctx,
		cancel:      cancel,
		sender:      senderSvc.Init(serverAddr),
	}
}

func (a *Agent) Start() {
	log.Printf("[AGENT:%s] : Starting with %d workers", a.name, a.workers)

	// starts agent's op.
	for i := range a.workers {
		a.wg.Add(1)

		go a.worker(i)
	}

	log.Printf("[AGENT:%s] : Agent started", a.name)
}

func (a *Agent) Stop() {
	log.Printf("[AGENT:%s] : Stopping Agent", a.name)

	// this signals for workers to stop
	a.cancel()

	// closes the queue that contains the jobs (agent will no more work on any new jobs)
	close(a.jobQueue)

	// this is required for every agent to complete it's running process
	a.wg.Wait()

	log.Printf("[AGENT:%s] : Agent stopped", a.name)
}

func (a *Agent) SubmitTransfer(source string) (string, error) {
	// create a new job that'll be added to queue
	job := models.NewTransferJob(source)

	// add the job ID to the memory pointing to the actual created just now
	a.statusMutex.Lock()
	a.statusStore[job.ID] = job
	a.statusMutex.Unlock()

	// add the created job to the queue from where the agent the is extracting and executing them
	select {
	case a.jobQueue <- job:
		log.Printf("[AGENT:%s] : Submitted transfer %s: %s", a.name, job.ID, source)

		return job.ID, nil
	case <-a.ctx.Done():
		return "", fmt.Errorf("agent is shutting down")
	}
}

func (a *Agent) RetrieveJob(id string) (*models.TransferJob, error) {
	// the resource might be accessed concurrently that's why locks are required
	a.statusMutex.RLock()
	defer a.statusMutex.Unlock()

	job, exists := a.statusStore[id]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", id)
	}

	return job, nil
}

func (a *Agent) ListJobs() []*models.TransferJob {
	// the resource might be accessed concurrently that's why locks are required
	a.statusMutex.RLock()
	defer a.statusMutex.Unlock()

	jobs := make([]*models.TransferJob, 0, len(a.statusStore))

	for _, job := range a.statusStore {
		jobs = append(jobs, job)
	}

	return jobs
}

func (a *Agent) worker(id int) {
	// mark the job as done post completion
	defer a.wg.Done()

	log.Printf("[AGENT:%s] : Worker %d started", a.name, id)

	for {
		select {
		case job, ok := <-a.jobQueue:
			if !ok {
				log.Printf("[AGENT:%s] : Worker %d stopped (queue closed/completed)", a.name, id)

				return
			}

			// process the job
			a.processJob(id, job)

		case <-a.ctx.Done():
			log.Printf("[AGENT:%s] : Worker %d stopped (Shutdown signal)", a.name, id)

			return
		}
	}
}

func (a *Agent) processJob(workerID int, job *models.TransferJob) {
	log.Printf("[AGENT:%s] : Worker %d processing job %s", a.name, workerID, job.ID)

	currentTime := time.Now()

	// move the status of the current job to be in progress
	a.updateJobStatus(job.ID, func(job *models.TransferJob) {
		job.Status = models.StatusInProgress
		job.StartedAt = &currentTime
	})

	// call the transfer for current source file
	err := a.sender.Send(job.SourcePath)

	completionTime := time.Now()

	if err != nil {
		log.Printf("[AGENT:%s] : Worker %d failed job %s: %v", a.name, workerID, job.ID, err)

		a.updateJobStatus(job.ID, func(job *models.TransferJob) {
			job.Status = models.StatusFailed
			job.Error = err.Error()
			job.StartedAt = &completionTime
		})
	} else {
		log.Printf("[AGENT:%s] : Worker %d completed job %s", a.name, workerID, job.ID)

		a.updateJobStatus(job.ID, func(job *models.TransferJob) {
			job.Status = models.StatusCompleted
			job.Error = err.Error()
			job.StartedAt = &completionTime
		})
	}

}

func (a *Agent) updateJobStatus(jobID string, fn func(*models.TransferJob)) {
	// the resource might be accessed concurrently that's why locks are required
	a.statusMutex.RLock()
	defer a.statusMutex.Unlock()

	// retrieve the job and call the update for that job
	// there is no status preservation and everything is done is memory and once memory is released ervery single resource will released
	job, exists := a.statusStore[jobID]
	if exists {
		fn(job)
	}
}
