package slack

import (
	"github.com/slack-go/slack/socketmode"
)

func (sc slackClient) EventProcessor() {
	for evt := range sc.client.Events {
		switch evt.Type {
		case socketmode.EventTypeConnecting:
			sc.logger.Info("Connecting to Slack with Socket Mode...")
		case socketmode.EventTypeConnectionError:
			sc.logger.Info("Connection failed. Retrying later...")
		case socketmode.EventTypeConnected:
			sc.logger.Info("Connected to Slack with Socket Mode.")
		}
	}
}