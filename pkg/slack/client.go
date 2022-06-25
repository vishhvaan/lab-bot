package slack

import (
	"fmt"
	stdlog "log"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	goslack "github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"

	"github.com/vishhvaan/lab-bot/pkg/config"
	"github.com/vishhvaan/lab-bot/pkg/jobs"
	"github.com/vishhvaan/lab-bot/pkg/logging"
)

type slackClient struct {
	api    *goslack.Client
	client *socketmode.Client
	slackBot
	logger    *log.Entry
	members   map[string]config.Member
	responses map[string]cb
	jobs      *jobs.JobHandler
}

type slackBot struct {
	bot        *goslack.Bot
	botID      string
	botChannel string
}

func CreateClient(secrets map[string]string, members map[string]config.Member, botChannel string) (sc *slackClient) {
	logFolder := logging.CreateLogFolder()
	logFileInternal := logging.CreateLogFile(logFolder, "slack_internal")
	slackLogger := logging.CreateNewLogger("slack", "slack")

	api := goslack.New(
		secrets["slack-bot-token"],
		goslack.OptionDebug(true),
		goslack.OptionLog(stdlog.New(logFileInternal,
			"api: ",
			stdlog.Lshortfile|stdlog.LstdFlags,
		)),
		goslack.OptionAppLevelToken(secrets["slack-app-token"]),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(stdlog.New(logFileInternal, "socketmode: ", stdlog.Lshortfile|stdlog.LstdFlags)),
	)

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

	r, err := api.GetConversationHistory(&goslack.GetConversationHistoryParameters{
		ChannelID: botChannelID,
		Limit:     1,
	})
	if err != nil {
		slackLogger.WithField("err", err).Fatal("Couldn't get history of message on the lab bot channel.")
	}
	botID := r.Messages[0].BotID
	bot, err := api.GetBotInfo(botID)
	if err != nil {
		slackLogger.WithField("err", err).Fatal("Couldn't find bot with ID.")
	}
	slackLogger.Info(fmt.Sprintf("%s has a bot ID of %s", bot.Name, botID))

	sc = &slackClient{
		api:    api,
		client: client,
		slackBot: slackBot{
			bot:        bot,
			botID:      botID,
			botChannel: botChannelID,
		},
		logger:    slackLogger,
		members:   members,
		responses: getResponses(),
		jobs:      jobs.CreateHandler(),
	}
	slackLogger.Info("Created Slack client.")
	return sc
}
