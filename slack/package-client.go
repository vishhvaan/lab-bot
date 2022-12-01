package slack

import (
	"github.com/vishhvaan/lab-bot/config"
)

var packageSlackClient *slackClient

func CreatePackageClient(secrets map[string]string, members map[string]config.Member, botChannel string) {
	packageSlackClient = CreateClient("global", secrets, members, botChannel)
}

func EventProcessor() {
	packageSlackClient.EventProcessor()
}

func RunSocketMode() {
	packageSlackClient.client.Run()
}

func React(timestamp string, channelID string, text string) error {
	return packageSlackClient.React(timestamp, channelID, text)
}

func Message(text string) (timestamp string, err error) {
	return packageSlackClient.Message(text)
}

func SendMessage(channel string, text string) (timestamp string, err error) {
	return packageSlackClient.SendMessage(channel, text)
}

func PostMessage(channelID string, text string) (timestamp string, err error) {
	return packageSlackClient.PostMessage(channelID, text)
}

func DeleteMessage(channelID string, timestamp string) error {
	return packageSlackClient.DeleteMessage(channelID, timestamp)
}

func UploadFile(channelID string, filePath string, title string) error {
	return packageSlackClient.UploadFile(channelID, filePath, title)
}

func ModifyMessage(channelID string, timestamp string, text string) error {
	return packageSlackClient.ModifyMessage(channelID, timestamp, text)
}

func PinMessage(channelID string, timestamp string) error {
	return packageSlackClient.PinMessage(channelID, timestamp)
}

func CommandStreamer(command string, outputType string, channelID string, timeout int) (output []string, err error) {
	return packageSlackClient.CommandStreamer(command, outputType, channelID, timeout)
}
