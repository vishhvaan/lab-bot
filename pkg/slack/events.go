package slack

import (
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func (sc *slackClient) EventProcessor() {
	for evt := range sc.client.Events {
		switch evt.Type {
		case socketmode.EventTypeConnecting:
			sc.logger.Info("Connecting to Slack with Socket Mode...")
		case socketmode.EventTypeConnectionError:
			sc.logger.Info("Connection failed. Retrying later...")
		case socketmode.EventTypeConnected:
			sc.logger.Info("Connected to Slack with Socket Mode.")
		case socketmode.EventTypeEventsAPI:
			eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
			if !ok {
				sc.logger.WithField("event", evt).Warn("Ignored event.")
				continue
			}
			// sc.logger.WithField("event", eventsAPIEvent).Info("Event recieved.")
			sc.client.Ack(*evt.Request)

			switch eventsAPIEvent.Type {
			case slackevents.CallbackEvent:
				go sc.cbEventProcessor(eventsAPIEvent)
			default:
				sc.logger.WithField("event", eventsAPIEvent).Warn(
					"Unsupported Events API event received.")
			}
		}
	}
}

func (sc *slackClient) cbEventProcessor(eventsAPIEvent slackevents.EventsAPIEvent) {
	innerEvent := eventsAPIEvent.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		sc.appMentionSubprocessor(ev)
	}
}

func (sc *slackClient) appMentionSubprocessor(ev *slackevents.AppMentionEvent) {
	go sc.logger.WithFields(log.Fields{
		"text":    ev.Text,
		"channel": sc.getChannelName(ev.Channel),
		"user":    sc.getUserName(ev.User),
	}).Info("App mentioned.")
	sc.launchCB(ev)
}
