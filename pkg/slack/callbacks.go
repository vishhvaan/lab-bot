package slack

import (
	"math/rand"
	"strings"

	"github.com/slack-go/slack/slackevents"

	"github.com/vishhvaan/lab-bot/pkg/functions"
)

func (sc *slackClient) commandInterpreter(ev *slackevents.AppMentionEvent) {
	noUID := strings.ReplaceAll(ev.Text, "<@"+sc.bot.UserID+">", "")
	fields := strings.Fields(noUID)
	if len(fields) == 0 {
		sc.logger.Info("Bot simply mentioned, responding with hello")
		hello(sc, ev, []string{""})
	} else {
		command := strings.ToLower(fields[0])
		if functions.Contains(functions.GetKeys(basicResponses), command) {
			f := basicResponses[command]
			f(sc, ev, fields)
		} else {
			sc.commander <- CommandInfo{
				Fields:    fields,
				Channel:   ev.Channel,
				TimeStamp: ev.TimeStamp,
			}
		}
	}
}

type cb func(sc *slackClient, ev *slackevents.AppMentionEvent, fields []string)

var basicResponses = map[string]cb{
	"hello": hello, "hai": hello, "hey": hello,
	"sup": hello, "hi": hello,
	"bye": bye, "goodbye": bye, "tata": bye,
	"sysinfo": sysinfo,
	"thanks":  thanks, "thank": thanks,
}

func hello(sc *slackClient, ev *slackevents.AppMentionEvent, fields []string) {
	response := "Hello, " + sc.getUserName(ev.User) + "! :party_parrot:"
	sc.PostMessage(MessageInfo{
		ChannelID: ev.Channel,
		Text:      response,
	})
}

func bye(sc *slackClient, ev *slackevents.AppMentionEvent, fields []string) {
	response := "Goodbye, " + sc.getUserName(ev.User) + "! :wave:"
	sc.PostMessage(MessageInfo{
		ChannelID: ev.Channel,
		Text:      response,
	})
}

func sysinfo(sc *slackClient, ev *slackevents.AppMentionEvent, fields []string) {
	response := functions.GetSysInfo()
	sc.PostMessage(MessageInfo{
		ChannelID: ev.Channel,
		Text:      response,
	})
}

func thanks(sc *slackClient, ev *slackevents.AppMentionEvent, fields []string) {
	allResponses := []string{
		"No problemo",
		"May the force be with you",
		":sunglasses:",
		":mechanical_arm:",
		":meow_party:",
		":meow_code:",
	}
	response := allResponses[rand.Intn(len(allResponses))]
	sc.PostMessage(MessageInfo{
		ChannelID: ev.Channel,
		Text:      response,
	})
}
