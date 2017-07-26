package shutdown

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

func noopShutdowner(ctx context.Context) {}
func TestAdd(t *testing.T) {
	lenBefore := len(shutdowners)
	Add(noopShutdowner)
	lenAfter := len(shutdowners)

	if lenBefore+1 != lenAfter {
		t.Error("Shutdowner wasn't added")
	}
}

type fakeSignalWaiter struct{}

func (f fakeSignalWaiter) wait() os.Signal {
	return syscall.SIGINT
}
func TestWaitWithNoopShutdowner(t *testing.T) {
	sw = fakeSignalWaiter{}
	Add(noopShutdowner)

	Wait()
}

func TestWaitWithLongRunningShutdowner(t *testing.T) {
	Timeout = 10 * time.Millisecond
	shutdowner := func(ctx context.Context) {
		ch := make(chan struct{})

		go func() {
			time.Sleep(10 * time.Second)
			close(ch)
		}()

		canceled := false
		timedout := false

		select {
		case <-ch:
			// Unreachable. I hope.
		case <-ctx.Done():
			// Context was canceled!
			canceled = true
		}

		if ctx.Err() == context.DeadlineExceeded {
			timedout = true
		}

		if !canceled {
			t.Error("Long running shutdowner wasn't cancelled")
		}

		if !timedout {
			t.Errorf("Long running shutdowner wasn't timed out: %v", ctx.Err())
		}
	}

	Add(shutdowner)
	Wait()
}
