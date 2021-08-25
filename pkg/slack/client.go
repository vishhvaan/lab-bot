package slack

import (
	"log"
	"os"

	goslack "github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func CreateClient(secrets map[string]string) (api *goslack.Client, client *socketmode.Client) {
	api = goslack.New(
		secrets["SLACK_BOT_TOKEN"],
		goslack.OptionDebug(true),
		goslack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		goslack.OptionAppLevelToken(secrets["SLACK_APP_TOKEN"]),
	)

	client = socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	return api, client
}
