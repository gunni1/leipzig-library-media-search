package command

import (
	"errors"
	"strconv"
	"strings"

	libClient "github.com/gunni1/leipzig-library-game-stock-api/library-le"
	tele "gopkg.in/telebot.v3"
)

// Listet alle Videospiele einer bestimmten Platform die aktuell in einer Zweigstelle Ausleihbar sind.
func ListBranchPlattformCommand(ctx tele.Context) error {
	client := libClient.Client{}
	platform, branchCode, argsErr := parsePlatformAndBranch(ctx.Args())
	if argsErr != nil {
		return ctx.Reply(argsErr.Error())
	}

	games := client.FindAvailabelGames(branchCode, platform)
	//TODO: func zum Umwandeln der Games-Liste in den Ergebnistext. Berücksichtigen, wenn Ergebniss leer ist!
	var replyBuilder strings.Builder
	//replyBuilder.WriteString(fmt.Sprintf("Spiele für %s in %s:\n", platform, libClient.BranchCodes[branchCode]))
	for _, game := range games {
		replyBuilder.WriteString(game.Title)
		replyBuilder.WriteString("\n")

	}
	return ctx.Send(replyBuilder.String())
}

func WelcomeCommand(ctx tele.Context) error {
	var replyBuilder strings.Builder

	return ctx.Send(replyBuilder.String())
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
