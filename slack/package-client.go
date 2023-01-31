package slack

var packageSlackClient *slackClient

func CreatePackageClient(botChannel string) {
	packageSlackClient = CreateClient("global", botChannel)
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

func ListPins(channelID string) (pinnedMessages map[string]string, err error) {
	return packageSlackClient.ListPins(channelID)
}

func PinMessage(channelID string, timestamp string) error {
	return packageSlackClient.PinMessage(channelID, timestamp)
}

func CommandStreamer(command string, outputType string, channelID string, timeout int) (output []string, err error) {
	return packageSlackClient.CommandStreamer(command, outputType, channelID, timeout)
}
