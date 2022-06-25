package slack

import (
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
	goslack "github.com/slack-go/slack"
)

type MessageInfo struct {
	Channel   string
	ChannelID string
	Text      string
}

func (sc *slackClient) MessageProcessor(m chan MessageInfo) {
	for message := range m {
		sc.Message(message)
	}
}

func (sc *slackClient) Message(i MessageInfo) (err error) {
	if i.Channel == "" && i.ChannelID != "" {
		sc.PostMessage(i)
		return nil
	} else if i.Channel != "" && i.ChannelID == "" {
		sc.SendMessage(i)
		return nil
	} else {
		return errors.New("need channel or channelID to send message")
	}
}

func (sc *slackClient) SendMessage(i MessageInfo) {
	_, _, _, err := sc.api.SendMessage(i.Channel, goslack.MsgOptionText(i.Text, false))
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't send message on Slack.")
	} else {
		sc.logger.WithFields(log.Fields{
			"text":    i.Text,
			"channel": i.Channel,
		}).Info("Sent message to Slack.")
	}
}

func (sc *slackClient) PostMessage(i MessageInfo) {
	_, _, err := sc.api.PostMessage(i.ChannelID, goslack.MsgOptionText(i.Text, false))
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't send message on Slack.")
	} else {
		sc.logger.WithFields(log.Fields{
			"text":    i.Text,
			"channel": sc.getChannelName(i.ChannelID),
		}).Info("Sent message to Slack.")
	}
}

func (sc *slackClient) textMatcher(message string) (match string, err string) {
	message = strings.ToLower(message)
	match = ""
	err = "no match found"
	for m := range sc.responses {
		if strings.Contains(message, m) {
			if match == "" {
				match = m
				err = ""
			} else {
				return "", "multiple matches found"
			}
		}
	}
	return match, err
}

func OnOffDetector(message string) (detected string) {
	lm := strings.ToLower(message)
	if strings.Contains(lm, " on") && !strings.Contains(lm, " off") {
		return "on"
	} else if strings.Contains(lm, " off") && !strings.Contains(lm, " on") {
		return "off"
	} else {
		return "both"
	}
}
