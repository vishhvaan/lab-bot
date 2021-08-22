package main

import (
	"flag"

	"github.com/TheoryDivision/lab-bot/pkg/config"
	"github.com/TheoryDivision/lab-bot/pkg/files"
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

	members := config.ParseMembers(membersFile)
	secrets := config.ParseSecrets(secretsFile)
}
