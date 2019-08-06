package balancer

import (
	"container/heap"
	"fmt"
)

const nRequester = 100
const nWorker = 10

type Request struct {
	fn func(in int) float64 // The operation to perform.
	in int                  // The  fn input
	c  chan float64         // The channel to return the result.
}

type Worker struct {
	requests chan Request // work to do (buffered channel)
	pending  int          // count of pending tasks
	index    int          // index in the heap
}

func (w *Worker) work(done chan *Worker) {
	for {
		req := <-w.requests     // get Request from balancer
		req.c <- req.fn(req.in) // call fn and send result
		done <- w               // we've finished this request
	}
}

type Pool []*Worker

func (p Pool) Len() int { return len(p) }

func (p Pool) Less(i, j int) bool {
	return p[i].pending < p[j].pending
}

func (p *Pool) Swap(i, j int) {
	a := *p
	a[i], a[j] = a[j], a[i]
	a[i].index = i
	a[j].index = j
}

func (p *Pool) Push(x interface{}) {
	n := len(*p)
	item := x.(*Worker)
	item.index = n
	*p = append(*p, item)
}

func (p *Pool) Pop() interface{} {
	old := *p
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*p = old[0 : n-1]
	return item
}

type Balancer struct {
	pool Pool
	done chan *Worker
}

func InitBalancer() *Balancer {
	done := make(chan *Worker, nWorker)
	// create nWorker WOK channels
	b := &Balancer{make(Pool, 0, nWorker), done}
	for i := 0; i < nWorker; i++ {
		w := &Worker{requests: make(chan Request, nRequester)}
		// put them in heap
		heap.Push(&b.pool, w)
		go w.work(b.done)
	}
	return b
}

func (b *Balancer) balance(work chan Request) {
	for {
		select {
		case req := <-work: // received a Request...
			b.dispatch(req) // ...so send it to a Worker
		case w := <-b.done: // a worker has finished ...
			b.completed(w)  // ...so update its info
		}
	}
}

// Send Request to worker
func (b *Balancer) dispatch(req Request) {
	// Grab the least loaded worker...
	w := heap.Pop(&b.pool).(*Worker)
	// ...send it the task.
	w.requests <- req
	// One more in its work queue.
	w.pending++
	// Put it into its place on the heap.
	heap.Push(&b.pool, w)
}

// Job is complete; update heap
func (b *Balancer) completed(w *Worker) {
	// One fewer in the queue.
	w.pending--
	// Remove it from heap.
	heap.Remove(&b.pool, w.index)
	// Put it into its place on the heap.
	heap.Push(&b.pool, w)
}

func (b *Balancer) print() {
	sum := 0
	sumsq := 0
	// Print pending stats for each worker
	for _, w := range b.pool {
		fmt.Printf("%d ", w.pending)
		sum += w.pending
		sumsq += w.pending * w.pending
	}
	// Print avg for worker pool
	avg := float64(sum) / float64(len(b.pool))
	variance := float64(sumsq)/float64(len(b.pool)) - avg*avg
	fmt.Printf(" %.2f %.2f\n", avg, variance)
}
