package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gunni1/leipzig-library-game-stock-api/domain"
	libClient "github.com/gunni1/leipzig-library-game-stock-api/library-le"
	tele "gopkg.in/telebot.v3"
)

type BotCommand struct {
	Prefix      string
	Description string
}

// Listet alle verfügbaren Switch-Spiele in einer bestimmten Zweigstelle.
func ListSwitchCommand(ctx tele.Context) error {
	client := libClient.Client{}
	if len(ctx.Args()) < 1 {
		return ctx.Reply("Bitte Zweigstelle angeben.")
	}
	branchArg := ctx.Args()[0]
	branchCode, pres := libClient.GetBranchCode(branchArg)
	if !pres {
		return ctx.Reply(fmt.Sprintf("Zweigstelle %s existiert nicht.", branchArg))
	}
	games := client.FindAvailabelGames(branchCode, "switch")
	reply := formatReply(games)
	return ctx.Send(reply)
}

// Listet alle Videospiele einer bestimmten Platform die aktuell in einer Zweigstelle Ausleihbar sind.
func ListBranchPlattformCommand(ctx tele.Context) error {
	client := libClient.Client{}
	platform, branchCode, argsErr := parsePlatformAndBranch(ctx.Args())
	if argsErr != nil {
		return ctx.Reply(argsErr.Error())
	}

	games := client.FindAvailabelGames(branchCode, platform)
	//TODO: func zum Umwandeln der Games-Liste in den Ergebnistext. Berücksichtigen, wenn Ergebniss leer ist!

	//replyBuilder.WriteString(fmt.Sprintf("Spiele für %s in %s:\n", platform, libClient.BranchCodes[branchCode]))
	reply := formatReply(games)
	return ctx.Send(reply)
}

func WelcomeCommand(ctx tele.Context) error {
	var replyBuilder strings.Builder
	replyBuilder.WriteString("Hi")
	return ctx.Send(replyBuilder.String())
}

// Erzeugt eine formatierte Ausgabe einer Liste von Titeln oder eine entsprechene Rückgabe bei leerer Liste.
func formatReply(games []domain.Game) string {
	if len(games) == 0 {
		return "Es wurden keine ausleihbaren Titel gefunden."
	}
	var replyBuilder strings.Builder
	for _, game := range games {
		replyBuilder.WriteString(game.Title)
		replyBuilder.WriteString("\n")
	}
	return replyBuilder.String()
}

// Holt aus den Command-Args die Platform und die Zweigstelle, oder liefert einen Fehlertext.
func parsePlatformAndBranch(args []string) (string, int, error) {
	if len(args) < 2 {
		return "", -1, errors.New("Bitte Plattform und Zweigstelle angeben. ")
	}
	platform := args[0]
	branchCode, parseBranchErr := strconv.Atoi(args[1])
	if parseBranchErr != nil {
		return "", -1, errors.New("Bitte die Zweigstelle als Zahl angeben.")
	}
	return platform, branchCode, nil
}
