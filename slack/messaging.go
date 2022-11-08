package slack

import (
	"errors"

	log "github.com/sirupsen/logrus"
	goslack "github.com/slack-go/slack"
)

type MessageInfo struct {
	Type      string
	Timestamp string
	Channel   string
	ChannelID string
	Text      string
}

func (sc *slackClient) React(timestamp string, channelID string, text string) (err error) {
	if timestamp == "" {
		return errors.New("cannot react to nonexistent message")
	} else if channelID == "" {
		return errors.New("need channelID to react")
	} else if text == "" {
		return errors.New("need something to react with")
	} else {
		err := sc.api.AddReaction(text, goslack.NewRefToMessage(channelID, timestamp))
		if err != nil {
			sc.logger.WithField("err", err).Error("Couldn't react to message on Slack.")
		} else {
			sc.logger.WithFields(log.Fields{
				"text":    text,
				"channel": sc.getChannelName(channelID),
			}).Info("Sent reaction to Slack.")
		}
		return nil
	}
}

func (sc *slackClient) Message(text string) (timestamp string, err error) {
	timestamp, err = sc.PostMessage(sc.botChannelID, text)
	return timestamp, err
}

func (sc *slackClient) SendMessage(channel string, text string) (timestamp string, err error) {
	if text == "" {
		return "", errors.New("cannot send empty message")
	}
	_, timestamp, _, err = sc.api.SendMessage(channel, goslack.MsgOptionText(text, false))
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't send message on Slack.")
	} else {
		sc.logger.WithFields(log.Fields{
			"text":    text,
			"channel": channel,
		}).Info("Sent message to Slack.")
	}
	return timestamp, err
}

func (sc *slackClient) PostMessage(channelID string, text string) (timestamp string, err error) {
	_, timestamp, err = sc.api.PostMessage(channelID, goslack.MsgOptionText(text, false))
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't send message on Slack.")
	} else {
		sc.logger.WithFields(log.Fields{
			"text":    text,
			"channel": sc.getChannelName(channelID),
		}).Info("Sent message to Slack.")
	}
	return timestamp, err
}
