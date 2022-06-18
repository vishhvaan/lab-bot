package slack

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	goslack "github.com/slack-go/slack"
)

func (sc *slackClient) SendMessage(channel string, text string) {
	_, _, _, err := sc.api.SendMessage(channel, goslack.MsgOptionText(text, false))
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't send message on Slack.")
	} else {
		sc.logger.WithFields(log.Fields{
			"text":    text,
			"channel": channel,
		}).Info("Sent message to Slack.")
	}
}

func (sc *slackClient) PostMessage(channelID string, text string) {
	_, _, err := sc.api.PostMessage(channelID, goslack.MsgOptionText(text, false))
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't send message on Slack.")
	} else {
		sc.logger.WithFields(log.Fields{
			"text":    text,
			"channel": sc.getChannelName(channelID),
		}).Info("Sent message to Slack.")
	}
}

func (sc *slackClient) RunSocketMode() {
	sc.client.Run()
}

func (sc *slackClient) getChannelName(channelID string) (channel string) {
	ch, err := sc.api.GetConversationInfo(channelID, false)
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't find conversation info.")
	}
	return ch.Name
}

func (sc *slackClient) getUserName(userID string) (user string) {
	us, err := sc.api.GetUserInfo(userID)
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't find conversation info.")
	}
	return us.Profile.DisplayName
}

func (sc *slackClient) PostStartupMessage() {
	currentTime := time.Now()
	hn, err := os.Hostname()
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't find OS hostname.")
	}
	m := fmt.Sprintf("lab-bot launched at %s on %s", currentTime.Format("2008-01-02 12:35:07 Wednesday"), hn)
	sc.PostMessage(sc.botChannel, m)
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't post startup message")
	}
}
