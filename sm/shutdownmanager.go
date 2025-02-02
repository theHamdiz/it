// Package sm - Because even programs need a retirement plan
package sm

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ShutdownAction is like a todo list for your program's last moments
type ShutdownAction struct {
	Name     string                      // What we're trying to clean up
	Action   func(context.Context) error // The actual cleanup (good luck)
	Timeout  time.Duration               // How long before we give up
	Critical bool                        // Whether failing this will haunt us
}

// ShutdownManager is like a funeral director for your services
// Makes sure everything gets a proper goodbye
type ShutdownManager struct {
	ctx      context.Context    // The end times
	cancel   context.CancelFunc // The kill switch
	actions  []ShutdownAction   // The farewell tour
	signals  []os.Signal        // What makes us give up
	errChan  chan error         // Where we log our regrets
	doneChan chan struct{}      // The final curtain
}

// NewShutdownManager creates a new end-of-life counselor for your application
// Accepts custom signals, or uses the classics (SIGINT, SIGTERM) if you're basic
func NewShutdownManager(signals ...os.Signal) *ShutdownManager {
	if len(signals) == 0 {
		// For the indecisive
		signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ShutdownManager{
		ctx:      ctx,
		cancel:   cancel,
		actions:  make([]ShutdownAction, 0), // Empty promises
		signals:  signals,
		errChan:  make(chan error, 1), // Room for one last mistake
		doneChan: make(chan struct{}), // The light at the end
	}
}

// AddAction adds another item to your program's bucket list
func (sm *ShutdownManager) AddAction(
	name string,
	action func(context.Context) error,
	timeout time.Duration,
	critical bool,
) {
	sm.actions = append(sm.actions, ShutdownAction{
		Name:     name,
		Action:   action,
		Timeout:  timeout,
		Critical: critical, // No pressure
	})
}

// Start begins watching for the end
// Like a vulture, but more professional
func (sm *ShutdownManager) Start() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sm.signals...)

	go func() {
		defer close(sm.doneChan) // Close the curtains on our way out

		select {
		case <-sigChan:
			log.Println("Received shutdown signal. Time for the long goodbye...")
			if err := sm.executeAll(); err != nil {
				sm.errChan <- err // One last disappointment
			}
		case <-sm.ctx.Done():
			// Someone pulled the plug early
			return
		}
	}()
}

// executeAll runs through the shutdown checklist
// Like a todo list, but with more panic
func (sm *ShutdownManager) executeAll() error {
	for _, action := range sm.actions {
		log.Printf("Executing last wishes: %s", action.Name)

		actionCtx, cancel := context.WithTimeout(sm.ctx, action.Timeout)
		err := action.Action(actionCtx)
		cancel() // Clean up after ourselves, one last time

		if err != nil {
			if action.Critical {
				return fmt.Errorf("critical shutdown action %s failed: %w", action.Name, err)
			}
			log.Printf("Non-critical shutdown action %s failed: %v", action.Name, err)
		}
	}
	return nil
}

// Wait blocks until everything is done or something goes terribly wrong
// Like watching paint dry, but with more anxiety
func (sm *ShutdownManager) Wait() error {
	<-sm.doneChan

	select {
	case err := <-sm.errChan:
		return err
	default:
		return nil // A clean death, how rare
	}
}

// Close pulls the plug immediately
// For when you're tired of waiting for natural causes
func (sm *ShutdownManager) Close() {
	sm.cancel()
}
