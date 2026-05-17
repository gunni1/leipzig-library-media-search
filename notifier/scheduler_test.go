package notifier

import (
	"fmt"
	"testing"

	. "github.com/stretchr/testify/assert"
)

// fakeChecker lets tests control IsAvailable results.
type fakeChecker struct {
	available bool
	err       error
	calls     []Subscription
}

func (fc *fakeChecker) IsAvailable(sub Subscription) (bool, error) {
	fc.calls = append(fc.calls, sub)
	return fc.available, fc.err
}

// fakeSender records which subscriptions were notified.
type fakeSender struct {
	notified []Subscription
	err      error
}

func (fs *fakeSender) Send(sub Subscription) error {
	fs.notified = append(fs.notified, sub)
	return fs.err
}

func TestScheduler_RunOnce_NotifiesAndDeletesWhenAvailable(t *testing.T) {
	store, _ := NewSubscriptionStore(t.TempDir())
	store.Save(Subscription{Email: "a@b.com", Title: "Dune", Type: "movie"})

	checker := &fakeChecker{available: true}
	sender := &fakeSender{}
	scheduler := NewScheduler(store, checker, sender)

	err := scheduler.RunOnce()
	Nil(t, err)

	Equal(t, 1, len(sender.notified))
	Equal(t, "Dune", sender.notified[0].Title)

	remaining, _ := store.GetAll()
	Equal(t, 0, len(remaining)) // subscription deleted after notification
}

func TestScheduler_RunOnce_SkipsWhenNotAvailable(t *testing.T) {
	store, _ := NewSubscriptionStore(t.TempDir())
	store.Save(Subscription{Email: "a@b.com", Title: "Dune", Type: "movie"})

	checker := &fakeChecker{available: false}
	sender := &fakeSender{}
	scheduler := NewScheduler(store, checker, sender)

	err := scheduler.RunOnce()
	Nil(t, err)

	Equal(t, 0, len(sender.notified))
	remaining, _ := store.GetAll()
	Equal(t, 1, len(remaining)) // subscription preserved
}

func TestScheduler_RunOnce_ContinuesOnSendError(t *testing.T) {
	store, _ := NewSubscriptionStore(t.TempDir())
	store.Save(Subscription{Email: "a@b.com", Title: "Dune", Type: "movie"})
	store.Save(Subscription{Email: "b@c.com", Title: "Zelda", Type: "game"})

	checker := &fakeChecker{available: true}
	sender := &fakeSender{err: fmt.Errorf("smtp timeout")}
	scheduler := NewScheduler(store, checker, sender)

	err := scheduler.RunOnce()
	Nil(t, err) // RunOnce itself doesn't fail

	// Both were attempted
	Equal(t, 2, len(sender.notified))
	// Neither deleted because send failed
	remaining, _ := store.GetAll()
	Equal(t, 2, len(remaining))
}
