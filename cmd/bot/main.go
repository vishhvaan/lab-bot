package main

import (
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/pkg/config"
	"github.com/vishhvaan/lab-bot/pkg/files"
	"github.com/vishhvaan/lab-bot/pkg/logging"
	"github.com/vishhvaan/lab-bot/pkg/slack"
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
	fmt.Println("::: Lab Bot :::")
	log.Info("Program Starting...")
	flag.Parse()

	log.Info("Checking config files.")
	files.CheckFile(membersFile)
	files.CheckFile(secretsFile)

	log.Info("Loading config files.")
	// members := config.ParseMembers(membersFile)
	secrets := config.ParseSecrets(secretsFile)
	slack.CheckSecrets(secrets)

	slackClient := slack.CreateClient(secrets)
	go slackClient.EventProcessor()
	slackClient.RunSocketMode()
	// slackClient.SendMessage("lab-bot-channel", "Hello, world!")
}
