package slack

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

func CheckSecrets(secrets map[string]string) {
	if secrets["SLACK_APP_TOKEN"] == "" {
		log.Fatal("App tolken not found.")
	}

	if secrets["SLACK_BOT_TOKEN"] == "" {
		log.Fatal("Bot tolken not found.")
	}

	if !strings.HasPrefix(secrets["SLACK_APP_TOKEN"], "xapp-") {
		log.Fatal("SLACK_APP_TOKEN must have the prefix \"xapp-\".")
	}

	if !strings.HasPrefix(secrets["SLACK_BOT_TOKEN"], "xoxb-") {
		log.Fatal("SLACK_APP_TOKEN must have the prefix \"xoxb-\".")
	}
}
