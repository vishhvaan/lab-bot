package main

import (
	"flag"
	"os"

	"github.com/TheoryDivision/lab-bot/pkg/config"
	"github.com/TheoryDivision/lab-bot/pkg/files"
	"github.com/TheoryDivision/lab-bot/pkg/slack"
	"go.uber.org/zap"
)

var (
	membersFile string
	secretsFile string
)

func init() {
	flag.StringVar(&membersFile, "members", "members.yml", "Location of the members file")
	flag.StringVar(&secretsFile, "secrets", "secrets.yml", "Location of the secrets file")
}

func main() {
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	files.CheckFile(membersFile)
	files.CheckFile(secretsFile)

	// members := config.ParseMembers(membersFile)
	secrets := config.ParseSecrets(secretsFile)
	err := slack.CheckSecrets(secrets)
	if err != nil {
		zap.L.Fatal("slack secret is invalid",
			zap.Error(err),
		)
		os.Exit(1)
	}

	api, client := slack.CreateClient(secrets)

}
