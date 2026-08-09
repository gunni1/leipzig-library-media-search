package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gunni1/leipzig-library-media-search/notifier"
)

func main() {
	port          := flag.String("port", "8081", "Port for the notifier HTTP server")
	dataDir       := flag.String("data-dir", "data/notifier", "Directory for subscription persistence")
	libraryURL    := flag.String("library-url", "http://localhost:3000", "Base URL of the library web service")
	checkInterval := flag.Duration("check-interval", time.Hour, "How often to check availability")
	smtpHost      := flag.String("smtp-host", "", "SMTP server hostname")
	smtpPort      := flag.String("smtp-port", "587", "SMTP server port")
	smtpUser      := flag.String("smtp-user", "", "SMTP username")
	smtpPass      := flag.String("smtp-pass", os.Getenv("SMTP_PASSWORD"), "SMTP password (default: $SMTP_PASSWORD)")
	smtpFrom      := flag.String("smtp-from", "", "Sender email address")
	flag.Parse()

	if *smtpHost == "" || *smtpFrom == "" {
		log.Fatal("smtp-host and smtp-from are required")
	}

	store, err := notifier.NewSubscriptionStore(*dataDir)
	if err != nil {
		log.Fatalf("failed to initialise subscription store: %v", err)
	}

	emailCfg := notifier.EmailConfig{
		Host:     *smtpHost,
		Port:     *smtpPort,
		Username: *smtpUser,
		Password: *smtpPass,
		From:     *smtpFrom,
	}
	sender    := notifier.NewEmailSender(emailCfg)
	checker   := notifier.NewAvailabilityChecker(*libraryURL)
	scheduler := notifier.NewScheduler(store, checker, sender)

	done := make(chan struct{})
	go scheduler.Start(*checkInterval, done)

	mux := notifier.NewNotifierMux(store, scheduler)
	fmt.Printf("notifier listening on port: %s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, mux))
}
