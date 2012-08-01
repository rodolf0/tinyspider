package concurrent

type WorkerPool chan bool

// create a worker pool to submit jobs
func MakeWorkerPool(size uint) WorkerPool {
	var pool = make(chan bool, size)
	for len(pool) < cap(pool) {
		pool <- true
	}
	return pool
}

// Schedule a job to be run on a free worker, if no workers
// are available the caller will block until one becomes available
func (w WorkerPool) Schedule(job func(...interface{}), args ...interface{}) {
	<-w
	go func() {
		job(args...)
		w <- true
	}()
}

// Return the number of running jobs
func (w WorkerPool) Running() int {
	return cap(w) - len(w)
}
