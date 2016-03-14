package runner

import (
	"testing"
	"time"
)

func TestOk(t *testing.T) {
	const tmo = 5 * time.Second
	r := New(tmo)

	r.Add(createTask(), createTask(), createTask())

	if err := r.Start(); err != nil {
		t.Fatal("Task execution error")
	}
}

func TestTmo(t *testing.T) {
	const tmo = 2 * time.Second
	r := New(tmo)

	r.Add(createTask(), createTask(), createTask())

	if err := r.Start(); err != nil {
		switch err {
		case ErrTimeout:
			t.Log("Timeout as expected")
			return
		default:
			t.Fatal("Unexpected error")
		}

	}
	t.Fatal("No error returned")
}

func createTask() func(int) {
	return func(id int) {
		time.Sleep(time.Duration(id) * time.Second)
	}
}
