package shutdown

import (
	"context"
	"os"
	"syscall"
	"testing"
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
