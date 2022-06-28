package slack

import (
	"strings"

	"github.com/slack-go/slack/slackevents"

	"github.com/vishhvaan/lab-bot/pkg/functions"
)

func (sc *slackClient) commandInterpreter(ev *slackevents.AppMentionEvent) {
	noUID := strings.ReplaceAll(ev.Text, "<@"+sc.bot.UserID+">", "")
	fields := strings.Fields(noUID)
	if len(fields) == 0 {
		sc.logger.Info("Bot simply mentioned, responding with hello")
		hello(sc, ev, "")
	} else {
		command := strings.ToLower(fields[0])
		if contains(GetKeys(basicResponses), command)
	}
}

var basicResponses = map[string]cb{
	"hello": hello, "hai": hello, "hey": hello,
	"sup": hello, "hi": hello,
	"bye": bye, "goodbye": bye, "tata": bye,
	"sysinfo": sysinfo,
}

type cb func(*slackClient, *slackevents.AppMentionEvent, string)

func (sc *slackClient) launchBasicCB(ev *slackevents.AppMentionEvent) {
	match, err := TextMatcher(ev.Text, GetKeys(responses))
	if err == nil {
		f := responses[match]
		f(sc, ev, match)
	} else if err.Error() == "no match found" {
		sc.logger.WithField("err", err).Warn("No callback function found.")
		sc.PostMessage(MessageInfo{
			ChannelID: ev.Channel,
			Text:      "I'm not sure what you sayin",
		})
	} else {
		sc.logger.WithField("err", err).Warn("Many callback functions found.")
		sc.PostMessage(MessageInfo{
			ChannelID: ev.Channel,
			Text:      "I can respond in multiple ways ...",
		})
	}
}

func hello(sc *slackClient, ev *slackevents.AppMentionEvent, match string) {
	response := "Hello, " + sc.getUserName(ev.User) + "! :party_parrot:"
	sc.PostMessage(MessageInfo{
		ChannelID: ev.Channel,
		Text:      response,
	})
}

func bye(sc *slackClient, ev *slackevents.AppMentionEvent, match string) {
	response := "Goodbye, " + sc.getUserName(ev.User) + "! :wave:"
	sc.PostMessage(MessageInfo{
		ChannelID: ev.Channel,
		Text:      response,
	})
}

func sysinfo(sc *slackClient, ev *slackevents.AppMentionEvent, match string) {
	response := functions.GetSysInfo()
	sc.PostMessage(MessageInfo{
		ChannelID: ev.Channel,
		Text:      response,
	})
}
