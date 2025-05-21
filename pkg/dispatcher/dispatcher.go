package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	worker "github.com/ameypant13/Record-to-Record-Synchronization-Service/pkg/sync-worker" // <-- Replace with your actual module path
	"golang.org/x/time/rate"
)

type Dispatcher struct {
	Limiter  *rate.Limiter
	JobChan  chan worker.SyncJob
	Worker   *worker.SyncWorker
	Wg       sync.WaitGroup
	quitChan chan struct{}
}

func NewDispatcher(syncWorker *worker.SyncWorker, rateLimit rate.Limit, burst int, buffer int) *Dispatcher {
	d := &Dispatcher{
		Limiter:  rate.NewLimiter(rateLimit, burst),
		JobChan:  make(chan worker.SyncJob, buffer),
		Worker:   syncWorker,
		quitChan: make(chan struct{}),
	}
	go d.loop()
	return d
}

// loop processes jobs from the channel and rate-limits actual API calls.
func (d *Dispatcher) loop() {
	for {
		select {
		case job := <-d.JobChan:
			err := d.Limiter.Wait(context.Background())
			if err != nil {
				fmt.Printf("Dispatcher limiter closed: %v\n", err)
				continue
			}
			d.Wg.Add(1)
			go func(j worker.SyncJob) {
				defer d.Wg.Done()
				res := d.Worker.ProcessJob(j)
				logResult(j, res)
			}(job)
		case <-d.quitChan:
			return
		}
	}
}

// Submit adds a job to the dispatcher's queue.
func (d *Dispatcher) Submit(job worker.SyncJob) error {
	select {
	case d.JobChan <- job:
		return nil
	default:
		return fmt.Errorf("dispatcher queue is full")
	}
}

// Shutdown gracefully shuts down dispatcher, waits for in-flight jobs.
func (d *Dispatcher) Shutdown() {
	close(d.quitChan)
	d.Wg.Wait()
}

func logResult(job worker.SyncJob, res worker.SyncResult) {
	if res.Status == "success" {
		fmt.Printf("[DISPATCH] Job [%s]: %s (%s)\n", job.Name, res.Status, res.Detail)
		enc, _ := json.MarshalIndent(res.Transformed, "  ", "  ")
		fmt.Println("  Output:", string(enc))
	} else {
		fmt.Printf("[NOT DISPATCH] Job [%s]: %s (%s)\n", job.Name, res.Status, res.Detail)
	}
}
