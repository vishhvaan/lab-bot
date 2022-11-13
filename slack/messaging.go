package slack

import (
	"bufio"
	"errors"
	"io"
	"os/exec"
	"time"

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
				"text":      text,
				"channelID": channelID,
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
			"text":      text,
			"channelID": channelID,
		}).Info("Sent message to Slack.")
	}
	return timestamp, err
}

func (sc *slackClient) DeleteMessage(channelID string, timestamp string) (err error) {
	_, _, err = sc.api.DeleteMessage(channelID, timestamp)
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't delete message on Slack.")
	} else {
		sc.logger.WithFields(log.Fields{
			"timestamp": timestamp,
			"channelID": channelID,
		}).Info("Deleted message on Slack.")
	}
	return err
}

func (sc *slackClient) UploadFile(channelID string, filePath string, title string) (err error) {
	_, err = sc.api.UploadFile(goslack.FileUploadParameters{
		File:     filePath,
		Title:    title,
		Channels: []string{channelID},
	})
	if err != nil {
		sc.logger.WithField("err", err).Error("Couldn't uploaded file to Slack.")
	} else {
		sc.logger.WithFields(log.Fields{
			"channelID": channelID,
		}).Info("Uploaded file to Slack.")
	}
	return err
}

func (sc *slackClient) CommandStreamer(command string, outputType string, channelID string, timeout int) (output []string, err error) {
	// timeout in seconds
	// outputType is either "out" or "err"
	cmd := exec.Command("bash", "-c", command)

	var stdpipe io.ReadCloser
	if outputType == "out" {
		stdpipe, err = cmd.StdoutPipe()
	} else if outputType == "err" {
		stdpipe, err = cmd.StderrPipe()
	} else {
		errMsg := "Command streamer needs a correct output type"
		sc.logger.WithField("err", err).Error(errMsg)
		return output, errors.New(errMsg)
	}

	if err != nil {
		errMsg := "Cannot create standard pipe"
		sc.logger.WithField("err", err).Error(errMsg)
		return output, errors.New(errMsg)
	}

	scanner := bufio.NewScanner(stdpipe)
	go func() {
		for scanner.Scan() {
			outputLine := scanner.Text()
			go func() {
				ts, err := PostMessage(channelID, outputLine)
				if err == nil {
					time.Sleep(time.Duration(timeout) * time.Second)
					sc.DeleteMessage(channelID, ts)
				} else {
					sc.logger.WithFields(log.Fields{
						"err":     err,
						"command": command,
						"line":    outputLine,
					}).Error("Cannot post command output")
				}
			}()
			output = append(output, outputLine)
		}
	}()

	err = cmd.Start()
	if err != nil {
		errMsg := "Error starting Cmd"
		sc.logger.WithFields(log.Fields{
			"err":     err,
			"command": command,
		}).Error(errMsg)
		return output, errors.New(errMsg)
	}

	err = cmd.Wait()
	if err != nil {
		errMsg := "Error waiting for Cmd"
		sc.logger.WithFields(log.Fields{
			"err":     err,
			"command": command,
		}).Error(errMsg)
		return output, errors.New(errMsg)
	}

	return output, err
}
