package slack

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/config"
	"github.com/vishhvaan/lab-bot/functions"
)

func CheckSlackSecrets() {
	if !functions.Contains(functions.GetKeys(config.Secrets), "slack-app-token") {
		log.Fatal("App token not found. (key is slack-app-token)")
	}

	if !functions.Contains(functions.GetKeys(config.Secrets), "slack-bot-token") {
		log.Fatal("App token not found. (key is slack-bot-token)")
	}

	if !strings.HasPrefix(config.Secrets["slack-app-token"], "xapp-") {
		log.Fatal("slack-app-token must have the prefix \"xapp-\".")
	}

	if !strings.HasPrefix(config.Secrets["slack-bot-token"], "xoxb-") {
		log.Fatal("slack-bot-token must have the prefix \"xoxb-\".")
	}
}
