package slack

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

func CheckSecrets(secrets map[string]string) {
	if secrets["slack-app-token"] == "" {
		log.Fatal("App token not found.")
	}

	if secrets["slack-bot-token"] == "" {
		log.Fatal("Bot token not found.")
	}

	if !strings.HasPrefix(secrets["slack-app-token"], "xapp-") {
		log.Fatal("slack-app-token must have the prefix \"xapp-\".")
	}

	if !strings.HasPrefix(secrets["slack-bot-token"], "xoxb-") {
		log.Fatal("slack-bot-token must have the prefix \"xoxb-\".")
	}
}
