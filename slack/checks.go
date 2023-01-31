package slack

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/config"
)

func CheckSlackSecrets() {
	if config.Secrets["slack-app-token"] == "" {
		log.Fatal("App token not found.")
	}

	if config.Secrets["slack-bot-token"] == "" {
		log.Fatal("Bot token not found.")
	}

	if !strings.HasPrefix(config.Secrets["slack-app-token"], "xapp-") {
		log.Fatal("slack-app-token must have the prefix \"xapp-\".")
	}

	if !strings.HasPrefix(config.Secrets["slack-bot-token"], "xoxb-") {
		log.Fatal("slack-bot-token must have the prefix \"xoxb-\".")
	}
}
