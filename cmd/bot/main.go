package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"

	command "github.com/gunni1/leipzig-library-game-stock-api/bot"
)

func main() {
	token := parseEnvMandatory("BOT_TOKEN")

	bot, botErr := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if botErr != nil {
		log.Fatal(botErr)
		return
	}

	//welcomeCommand := command.BotCommand{Prefix: "/start", Description: "Zeigt die Liste aller Bot-Funktionen an."}

	bot.Handle("/start", command.WelcomeCommand)
	bot.Handle("/list", command.ListBranchPlattformCommand)

	go setupHealthEndpoint()
	log.Println("Bot Ready.")
	bot.Start()
}

func setupHealthEndpoint() {
	http.HandleFunc("/", func(writer http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(writer, "ready")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func parseEnvMandatory(variableKey string) string {
	variableValue := os.Getenv(variableKey)
	if variableValue == "" {
		log.Fatalln("Environment variable: " + variableKey + " is empty")
	}
	return variableValue
}
