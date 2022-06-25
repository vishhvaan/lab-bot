package main

import (
	"flag"
	"fmt"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/pkg/config"
	"github.com/vishhvaan/lab-bot/pkg/files"
	"github.com/vishhvaan/lab-bot/pkg/jobs"
	"github.com/vishhvaan/lab-bot/pkg/logging"
	"github.com/vishhvaan/lab-bot/pkg/slack"
)

var (
	membersFile string
	secretsFile string
	botName     string
	botChannel  string
)

func init() {
	logging.Setup()
	exePath := logging.FindExeDir()
	flag.StringVar(&membersFile, "members", path.Join(exePath, "members.yml"), "Location of the members file")
	flag.StringVar(&secretsFile, "secrets", path.Join(exePath, "secrets.yml"), "Location of the secrets file")
	flag.StringVar(&botChannel, "channel", "lab-bot-channel", "Name of the bot channel")
}

func main() {
	fmt.Println("::: Lab Bot :::")
	log.Info("Program Starting...")
	flag.Parse()

	log.Info("Checking config files.")
	files.CheckFile(membersFile)
	files.CheckFile(secretsFile)

	log.Info("Loading config files.")
	members := config.ParseMembers(membersFile)
	secrets := config.ParseSecrets(secretsFile)
	slack.CheckSecrets(secrets)

	messages := make(chan slack.MessageInfo)
	commands := make(chan slack.CommandInfo)

	slackClient := slack.CreateClient(secrets, members, botChannel, commands)
	go slackClient.MessageProcessor(messages)
	go slackClient.EventProcessor()
	go slackClient.RunSocketMode()

	jobHandler := jobs.CreateHandler(messages)
	jobHandler.CommandReciever(commands)
}
