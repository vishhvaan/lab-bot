package slack

import (
	"errors"
	"strings"
)

func CheckSecrets(secrets map[string]string) (err error) {
	if secrets["SLACK_APP_TOKEN"] == "" {
		return errors.New("app tolken not found")
	}

	if secrets["SLACK_BOT_TOKEN"] == "" {
		return errors.New("bot tolken not found")
	}

	if !strings.HasPrefix(secrets["SLACK_APP_TOKEN"], "xapp-") {
		return errors.New("SLACK_APP_TOKEN must have the prefix \"xapp-\"")
	}

	if !strings.HasPrefix(secrets["SLACK_BOT_TOKEN"], "xoxb-") {
		return errors.New("SLACK_APP_TOKEN must have the prefix \"xoxb-\"")
	}

	return nil
}
