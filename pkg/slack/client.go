package slack

import (
	stdlog "log"

	log "github.com/sirupsen/logrus"
	goslack "github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"

	"github.com/TheoryDivision/lab-bot/pkg/logging"
)

type slackClient struct {
	api    *goslack.Client
	client *socketmode.Client
	logger *log.Entry
}

func CreateClient(secrets map[string]string) (sc slackClient) {

	logFolder := logging.CreateLogFolder()
	logFileInternal := logging.CreateLogFile(logFolder, "slack_internal")
	slackLogger := logging.CreateNewLogger("slack", "slack")

	api := goslack.New(
		secrets["SLACK_BOT_TOKEN"],
		goslack.OptionDebug(true),
		goslack.OptionLog(stdlog.New(logFileInternal, "api: ", stdlog.Lshortfile|stdlog.LstdFlags)),
		goslack.OptionAppLevelToken(secrets["SLACK_APP_TOKEN"]),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(stdlog.New(logFileInternal, "socketmode: ", stdlog.Lshortfile|stdlog.LstdFlags)),
	)

	sc = slackClient{api: api, client: client, logger: slackLogger}
	slackLogger.Info("Created Slack client.")
	return sc
}
