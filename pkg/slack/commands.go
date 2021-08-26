package slack

import (
	log "github.com/sirupsen/logrus"
	goslack "github.com/slack-go/slack"
)

func (sc slackClient) SendMessage(channel string, text string) {
	_, _, _, err := sc.api.SendMessage("lab-bot-channel", goslack.MsgOptionText("hello, world!", false))
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't send message on Slack.")
	} else {
		sc.logger.WithFields(log.Fields{
			"text":    text,
			"channel": channel,
		}).Info("Sent message on slack.")
	}
}
