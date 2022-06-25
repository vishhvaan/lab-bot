package slack

import (
	"github.com/slack-go/slack/slackevents"

	"github.com/vishhvaan/lab-bot/pkg/functions"
)

type cb func(*slackClient, *slackevents.AppMentionEvent, string)

func getResponses() map[string]cb {
	return map[string]cb{
		"hello": hello, "hai": hello, "hey": hello,
		"sup": hello, "hi": hello,
		"bye": bye, "goodbye": bye, "tata": bye,
		"sysinfo": sysinfo,
	}
}

func (sc *slackClient) launchCB(ev *slackevents.AppMentionEvent) {
	match, err := sc.textMatcher(ev.Text)
	if err == "" {
		f := sc.responses[match]
		f(sc, ev, match)
	} else if err == "no match found" {
		sc.logger.Warn("No callback function found.")
		sc.PostMessage(ev.Channel, "I'm not sure what you sayin")
	} else {
		sc.logger.Warn("Many callback functions found.")
		sc.PostMessage(ev.Channel, "I can respond in multiple ways ...")
	}
}

func hello(sc *slackClient, ev *slackevents.AppMentionEvent, match string) {
	response := "Hello, " + sc.getUserName(ev.User) + "! :party_parrot:"
	sc.PostMessage(ev.Channel, response)
}

func bye(sc *slackClient, ev *slackevents.AppMentionEvent, match string) {
	response := "Goodbye, " + sc.getUserName(ev.User) + "! :wave:"
	sc.PostMessage(ev.Channel, response)
}

func sysinfo(sc *slackClient, ev *slackevents.AppMentionEvent, match string) {
	response := functions.GetSysInfo()
	sc.PostMessage(ev.Channel, response)
}
