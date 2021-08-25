package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/TheoryDivision/lab-bot/pkg/config"
	"github.com/TheoryDivision/lab-bot/pkg/files"
	"github.com/TheoryDivision/lab-bot/pkg/logging"
	"github.com/TheoryDivision/lab-bot/pkg/slack"
)

var (
	membersFile string
	secretsFile string
)

func init() {
	logging.Setup()
	flag.StringVar(&membersFile, "members", "members.yml", "Location of the members file")
	flag.StringVar(&secretsFile, "secrets", "secrets.yml", "Location of the secrets file")
}

func main() {
	flag.Parse()

	files.CheckFile(membersFile)
	files.CheckFile(secretsFile)

	// members := config.ParseMembers(membersFile)
	secrets := config.ParseSecrets(secretsFile)
	err := slack.CheckSecrets(secrets)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Slack secret is invalid.")
	}

	// api, client := slack.CreateClient(secrets)

}
