package executor

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// WorkQueue is the struct for the work queue
type WorkQueue struct {
	numWorkers            int
	pendingScriptWorkChan chan Script
	ResultsChan           chan ExecutionResult
	doneChan              chan interface{}
	Wg                    *sync.WaitGroup
}

// NewWorkQueue returns a new instance of WorkQueue
func NewWorkQueue(capacity int, maxWorkers int, totalScripts int) *WorkQueue {
	wq := &WorkQueue{
		numWorkers:            maxWorkers,
		pendingScriptWorkChan: make(chan Script, capacity),
		ResultsChan:           make(chan ExecutionResult, capacity),
		doneChan:              make(chan interface{}, 1),
		Wg:                    &sync.WaitGroup{},
	}
	return wq
}

// SubmitTask adds a new script execution task to the channel
func (w *WorkQueue) SubmitTask(script Script) {
	log.Debug("Submiting script ", script.Path, " to be executed...")
	w.Wg.Add(1)
	w.pendingScriptWorkChan <- script
	log.Debug("Script submitted")
}

// Process starts the workers that will process jobs
func (w *WorkQueue) Process() {
	for n := 1; n <= w.numWorkers; n++ {
		log.Debug("Starting Execution Worker #", n)
		go w.execWorker(n)
	}
}

func (w *WorkQueue) execWorker(id int) {
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case script := <-w.pendingScriptWorkChan:
			log.Debugf("[Worker #%d] Running script %s", id, script.Path)
			scriptResult := RunScript(script, 5)

			if scriptResult.Error != nil {
				log.Errorf("[Worker #%d] Encountered error executing script %s (Error: %v)", id, script.Name, scriptResult.Error)
			} else {
				log.Debugf("[Worker #%d] Script %s completed execution. Result: %v", id, script.Name, scriptResult)
			}
			w.ResultsChan <- scriptResult
		case <-w.doneChan:
			log.Debugf("[Worker #%d] Shutting down", id)
			return
		case <-ticker.C:
			log.Tracef("[Worker #%d] Waiting for work", id)
		}
	}
}
