package slack

import (
	"strings"

	"github.com/slack-go/slack/slackevents"
)

type cb func(*slackClient, *slackevents.AppMentionEvent, string)

func getResponses() map[string]cb {
	return map[string]cb{
		"hello": hello, "hai": hello, "hey": hello,
		"sup": hello, "hi": hello,
		"bye": bye, "goodbye": bye, "tata": bye,
		"coffee machine": coffee,
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

func (sc *slackClient) textMatcher(message string) (match string, err string) {
	message = strings.ToLower(message)
	match = ""
	err = "no match found"
	for m := range sc.responses {
		if strings.Contains(message, m) {
			if match == "" {
				match = m
				err = ""
			} else {
				return "", "multiple matches found"
			}
		}
	}
	return match, err
}


func onOffDetector(message string) (detected string) {
	lm := strings.ToLower(message)
	if strings.Contains(lm, " on") && !strings.Contains(lm, " off") {
		return "on"
	} else if strings.Contains(lm, " off") && !strings.Contains(lm, " on") {
		return "off"
	} else {
		return "both"
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

func coffee(sc *slackClient, ev *slackevents.AppMentionEvent, match string) {
	// control := onOffDetector(ev.Text)

	response := "WIP" + sc.getUserName(ev.User) + "! :wave:"
	sc.PostMessage(ev.Channel, response)
}