package main

import (
	"flag"
	"fmt"
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
}
