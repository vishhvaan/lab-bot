package main

import (
	"flag"
	"log"
	"os"

	"github.com/TheoryDivision/lab-bot/pkg/config"
	"github.com/TheoryDivision/lab-bot/pkg/files"
	"github.com/TheoryDivision/lab-bot/pkg/slack"
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
	files.CheckFile(membersFile)
	files.CheckFile(secretsFile)

	// members := config.ParseMembers(membersFile)
	secrets := config.ParseSecrets(secretsFile)
	err := slack.CheckSecrets(secrets)
	if err != nil {
		log.Fatalf("slack secret is invalid: %v", err)
		os.Exit(1)
	}
}
