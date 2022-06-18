package slack

import (
	"errors"
	stdlog "log"

	log "github.com/sirupsen/logrus"
	goslack "github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"

	"github.com/vishhvaan/lab-bot/pkg/config"
	"github.com/vishhvaan/lab-bot/pkg/logging"
)

type slackClient struct {
	api        *goslack.Client
	client     *socketmode.Client
	bot        *goslack.Bot
	botChannel string
	logger     *log.Entry
	members    map[string]config.Member
	responses  map[string]cb
}

func CreateClient(secrets map[string]string, members map[string]config.Member, botChannel string) (sc *slackClient) {
	logFolder := logging.CreateLogFolder()
	logFileInternal := logging.CreateLogFile(logFolder, "slack_internal")
	slackLogger := logging.CreateNewLogger("slack", "slack")

	api := goslack.New(
		secrets["SLACK_BOT_TOKEN"],
		goslack.OptionDebug(true),
		goslack.OptionLog(stdlog.New(logFileInternal,
			"api: ",
			stdlog.Lshortfile|stdlog.LstdFlags,
		)),
		goslack.OptionAppLevelToken(secrets["SLACK_APP_TOKEN"]),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(stdlog.New(logFileInternal, "socketmode: ", stdlog.Lshortfile|stdlog.LstdFlags)),
	)

	u, err := api.GetUserIdentity()
	if err != nil {
		slackLogger.WithField("err", err).Fatal("Couldn't find self on Slack")
	}
	bot, err := api.GetBotInfo(u.User.ID)
	if err != nil {
		slackLogger.WithField("err", err).Fatal("Couldn't find bot")
	}

	// verify that the botChannel exists
	botChannels, _, err := api.GetConversationsForUser(&goslack.GetConversationsForUserParameters{
		UserID:          bot.ID,
		ExcludeArchived: true,
	})
	if err != nil {
		slackLogger.WithField("err", err).Fatal("Couldn't get list of bot channels")
	}

	var botChannelID string
	err = errors.New("Couldn't find bot channel")
	for _, c := range botChannels {
		if c.Name == botChannel {
			slackLogger.Info("Found bot channel on Slack")
			botChannelID = c.ID
			err = nil
			break
		}
	}
	if err != nil {
		slackLogger.Fatal(err)
	}

	sc = &slackClient{
		api:        api,
		client:     client,
		bot:        bot,
		botChannel: botChannelID,
		logger:     slackLogger,
		members:    members,
		responses:  getResponses(),
	}
	slackLogger.Info("Created Slack client.")
	return sc
}
