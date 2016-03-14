package pool
import (
	"sync"
	"io"
	"errors"
	"log"
)

var ErrPoolClosed = errors.New("Pool has been closed.")

type Pool struct  {
	m sync.Mutex
	resources chan io.Closer
	factory func() (io.Closer, error)
	closed bool
}

func New(fn func() (io.Closer, error), size uint) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("Size value too small.")
	}

	return &Pool{
		factory: fn,
		resources: make(chan io.Closer, size),
	}, nil
}

// Acquire retrieves a resource from the pool
func (p *Pool) Acquire() (io.Closer, error) {
	select {
	// check for a free resource
	case r, ok := <-p.resources:
		log.Println("Acquire:", "Shared Resource")
		if !ok {
			return nil, ErrPoolClosed
		}
		return r, nil
	// Provide a new resuource since there are none available.
	default:
		log.Println("Acquire:", "New Resource")
		return p.factory()
	}
}

// Release places a new resuource onto the pool.
func (p *Pool) Release(r io.Closer) {
	// Secure this operation with the Close operation.
	p.m.Lock()
	defer p.m.Unlock()

	// If pool is closed discard the resource
	if p.closed {
		r.Close()
		return
	}

	select {
	// Attempt to place the new resource on the queue
	case p.resources <- r:
		log.Println("Release:",  "In Queue")
	default:
		log.Println("Release:", "Closing")
		r.Close()
	}
}

// Close will shutdown the pool and close all existing resources.
func (p *Pool) Close() {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return
	}

	p.closed = true

	// Close the channel before we drain the channel of its resources.
	// If we don't do this, we will have a deadlock.
	close(p.resources)

	// Close the resources
	for r := range p.resources {
		r.Close()
	}
}


