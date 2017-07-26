package shutdown

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	// Timeout ...
	Timeout time.Duration

	// Logger ...
	Logger LogPrinter

	names = map[os.Signal]string{
		syscall.SIGINT:  "SIGINT",
		syscall.SIGTERM: "SIGTERM",
	}

	shutdowners = []Shutdowner{}
	sw          signalWaiter
)

// Shutdowner represents convenient alias for user-defined callback
type Shutdowner func(ctx context.Context)

// Add func registers user-defined Shutdowner to be handled gracefully
func Add(s Shutdowner) {
	shutdowners = append(shutdowners, s)
}

// Wait for shutdown signal and handle it!
func Wait() {
	if Logger == nil {
		Logger = log.New(ioutil.Discard, "", 0)
	}

	if Timeout == 0 {
		Timeout = 5 * time.Second
	}

	if sw == nil {
		sw = osSignalWaiter{}
	}

	sig := sw.wait()
	Logger.Printf("Signal '%s' recieved.", names[sig])

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(shutdowners))

	for _, cb := range shutdowners {
		go func(s Shutdowner) {
			defer wg.Done()
			s(ctx)
		}(cb)
	}

	completed := make(chan struct{})
	go func() {
		defer close(completed)
		wg.Wait()
	}()

	select {
	case <-completed:
		Logger.Printf("All shutdowners are finished")
	case <-ctx.Done():
		Logger.Printf("Shutdowners was timed out: %v", ctx.Err())
	}
}

// LogPrinter ...
type LogPrinter interface {
	Printf(format string, v ...interface{})
}

func waitSignal() os.Signal {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return <-ch
}

type signalWaiter interface {
	wait() os.Signal
}

type osSignalWaiter struct{}

func (w osSignalWaiter) wait() os.Signal {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return <-ch
}
