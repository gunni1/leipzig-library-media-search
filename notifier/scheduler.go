package notifier

import (
	"log"
	"time"
)

// Checker checks whether a media item is currently available.
type Checker interface {
	IsAvailable(sub Subscription) (bool, error)
}

// Sender sends a notification for a subscription.
type Sender interface {
	Send(sub Subscription) error
}

// Scheduler periodically checks all subscriptions and sends notifications.
type Scheduler struct {
	store   *SubscriptionStore
	checker Checker
	sender  Sender
}

// NewScheduler creates a Scheduler wired to the given store, checker, and sender.
func NewScheduler(store *SubscriptionStore, checker Checker, sender Sender) *Scheduler {
	return &Scheduler{store: store, checker: checker, sender: sender}
}

// RunOnce checks all subscriptions once: notify and delete those whose item is available.
func (scheduler *Scheduler) RunOnce() error {
	subs, err := scheduler.store.GetAll()
	log.Printf("checking %d for subscriptions to notify...\n", len(subs))
	if err != nil {
		return err
	}
	for _, sub := range subs {
		available, checkErr := scheduler.checker.IsAvailable(sub)
		if checkErr != nil {
			log.Printf("scheduler: availability check failed for %q: %v", sub.Title, checkErr)
			continue
		}
		if !available {
			continue
		}
		if sendErr := scheduler.sender.Send(sub); sendErr != nil {
			log.Printf("scheduler: failed to notify %s for %q: %v", sub.Email, sub.Title, sendErr)
			continue
		}
		if deleteErr := scheduler.store.Delete(sub.ID); deleteErr != nil {
			log.Printf("scheduler: failed to delete subscription %s: %v", sub.ID, deleteErr)
		}
	}
	return nil
}

// Start runs RunOnce on every interval tick. Blocks until the done channel is closed.
func (scheduler *Scheduler) Start(interval time.Duration, done <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := scheduler.RunOnce(); err != nil {
				log.Printf("scheduler: run failed: %v", err)
			}
		case <-done:
			return
		}
	}
}
