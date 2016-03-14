package runner

import (
	"errors"
	"os"
	"os/signal"
	"time"
)

var ErrTimeout = errors.New("received timeout")
var ErrInterrupt = errors.New("received interrupt")

type Runner struct {
	interrupt chan os.Signal   // OS interrupt signals
	complete  chan error       // To report when processing is done
	timeout   <-chan time.Time // To report timeout
	tasks     []func(int)
}

func New(d time.Duration) *Runner {
	return &Runner{
		interrupt: make(chan os.Signal, 1), // must use buffered channel, otherwise signal might be lost
		complete:  make(chan error),
		timeout:   time.After(d),
	}
}

func (r *Runner) Add(tasks ...func(int)) {
	r.tasks = append(r.tasks, tasks...)
}

func (r *Runner) Start() error {
	// receive all interrupt based signals
	signal.Notify(r.interrupt, os.Interrupt)

	go func() {
		r.complete <- r.run()
	}()

	select {
	case err := <-r.complete:
		return err
	case <-r.timeout:
		return ErrTimeout
	}
}

func (r *Runner) run() error {
	for id, task := range r.tasks {
		// Check for interrupt from OS
		if r.gotInterrupt() {
			return ErrInterrupt
		}

		// Execute the registered task
		task(id)
	}
	return nil
}

func (r *Runner) gotInterrupt() bool {
	select {
	case <-r.interrupt:
		signal.Stop(r.interrupt)
		return true
	default:
		return false
	}
}
