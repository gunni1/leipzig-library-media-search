package notifier

import (
	"log"
	"net/http"
)

// NewNotifierMux creates the HTTP mux for the notifier service.
// scheduler is accepted for future use; handlers do not call it directly.
func NewNotifierMux(store *SubscriptionStore, scheduler *Scheduler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /subscriptions", makeSubscribeHandler(store))
	mux.HandleFunc("DELETE /subscriptions/{id}", makeDeleteHandler(store))
	return mux
}

func makeSubscribeHandler(store *SubscriptionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.PostFormValue("email")
		title := r.PostFormValue("title")
		mediaType := r.PostFormValue("type")
		platform := r.PostFormValue("platform")
		if email == "" || title == "" || mediaType == "" {
			http.Error(w, "email, title and type are required", http.StatusBadRequest)
			return
		}
		sub := Subscription{
			Email:    email,
			Title:    title,
			Type:     mediaType,
			Platform: platform,
		}
		if _, err := store.Save(sub); err != nil {
			log.Printf("notifier: failed to save subscription: %v", err)
			http.Error(w, "could not save subscription", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func makeDeleteHandler(store *SubscriptionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := store.Delete(id); err != nil {
			log.Printf("notifier: failed to delete subscription %s: %v", id, err)
			http.Error(w, "could not delete subscription", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
