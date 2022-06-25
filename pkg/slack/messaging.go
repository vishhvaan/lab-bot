package slack

import (
	"errors"
	"strings"

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

func (sc *slackClient) MessageProcessor(m chan MessageInfo) {
	var err error
	for message := range m {
		if message.Type == "react" {
			err = sc.React(message)
		} else {
			err = sc.Message(message)
		}
		if err != nil {
			sc.logger.Error(err)
		}
	}
}

func (sc *slackClient) React(i MessageInfo) (err error) {
	if i.Timestamp == "" {
		return errors.New("cannot react to nonexistent message")
	} else if i.ChannelID == "" {
		return errors.New("need channelID to react")
	} else if i.Text == "" {
		return errors.New("need something to react with")
	} else {
		sc.MsgReact(i)
		return nil
	}
}

func (sc *slackClient) Message(i MessageInfo) (err error) {
	if i.Text == "" {
		return errors.New("cannot send empty message")
	}

	if i.Channel == "" && i.ChannelID != "" {
		sc.PostMessage(i)
		return nil
	} else if i.Channel != "" && i.ChannelID == "" {
		sc.SendMessage(i)
		return nil
	} else {
		sc.PostMessage(MessageInfo{
			ChannelID: sc.botChannelID,
			Text:      i.Text,
		})
		return nil
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

func (sc *slackClient) MsgReact(i MessageInfo) {
	err := sc.api.AddReaction(i.Text, goslack.NewRefToMessage(i.ChannelID, i.Timestamp))
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't react to message on Slack.")
	} else {
		sc.logger.WithFields(log.Fields{
			"text":    i.Text,
			"channel": sc.getChannelName(i.ChannelID),
		}).Info("Sent reaction to Slack.")
	}
}

func TextMatcher(message string, possibilities []string) (match string, err error) {
	message = strings.ToLower(message)
	match = ""
	err = errors.New("no match found")
	for _, m := range possibilities {
		if strings.Contains(message, m) {
			if match == "" {
				match = m
				err = nil
			} else {
				return "", errors.New("multiple matches found")
			}
		}
	}
	return match, err
}

func GetKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
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
