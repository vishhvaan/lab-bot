package slack

import (
	"fmt"
	stdlog "log"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	goslack "github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"

	"github.com/vishhvaan/lab-bot/config"
	"github.com/vishhvaan/lab-bot/logging"
)

type slackClient struct {
	name   string
	api    *goslack.Client
	client *socketmode.Client
	logger *log.Entry
	slackBot
}

type slackBot struct {
	bot          *goslack.Bot
	botChannelID string
}

func CreateClient(name string, botChannel string) (sc *slackClient) {
	logFolder := logging.CreateLogFolder()
	logFileInternal := logging.CreateLogFile(logFolder, "slack_internal")
	slackLogger := logging.CreateNewLogger("slack", "slack")

	api := goslack.New(
		config.Secrets["slack-bot-token"],
		goslack.OptionDebug(true),
		goslack.OptionLog(stdlog.New(logFileInternal,
			"api: ",
			stdlog.Lshortfile|stdlog.LstdFlags,
		)),
		goslack.OptionAppLevelToken(config.Secrets["slack-app-token"]),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(stdlog.New(logFileInternal, "socketmode: ", stdlog.Lshortfile|stdlog.LstdFlags)),
	)

	r, err := api.AuthTest()
	if err != nil {
		slackLogger.WithField("err", err).Fatal("Slack connection test failed.")
	}
	bot, err := api.GetBotInfo(r.BotID)
	if err != nil {
		slackLogger.WithField("err", err).Fatal("Couldn't find bot with ID.")
	}
	slackLogger.Info(fmt.Sprintf("%s has a bot ID of %s", bot.Name, bot.ID))

	currentTime := time.Now()
	hn, err := os.Hostname()
	if err != nil {
		slackLogger.WithField("err", err).Error("Couldn't find OS hostname.")
	}
	m := fmt.Sprintf("lab-bot launched at %s on %s", currentTime.Format(time.UnixDate), hn)
	botChannelID, _, _, err := api.SendMessage(botChannel, goslack.MsgOptionText(m, false))
	if err != nil {
		slackLogger.WithField("err", err).Fatal("Couldn't send message to the lab bot channel.")
	} else {
		slackLogger.WithFields(log.Fields{
			"text":    m,
			"channel": botChannel,
		}).Info("Sent startup message to Slack.")
	}

	sc = &slackClient{
		name:   name,
		api:    api,
		client: client,
		logger: slackLogger,
		slackBot: slackBot{
			bot:          bot,
			botChannelID: botChannelID,
		},
	}

	slackLogger.Info("Created " + name + " Slack client.")
	return sc
}
