package jobs

import (
	// "github.com/go-co-op/gocron"

	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack/slackevents"

	"github.com/vishhvaan/lab-bot/pkg/logging"
	"github.com/vishhvaan/lab-bot/pkg/slack"
)

/*
Todo:
schedule jobs with github.com/go-co-op/gocron
keep track of status
struct for each job with parameters
jobs interface with shared commands
add ability to keep track of messageIDs
create processing for message emojis
task done with reactions -> ends interaction for that time
else post on the lab-bot-channel with the info
(or start group conv) -> ends interaction for that time

find members with particular roles
apply jobs to those roles

upload new members.yml via messages
check and rewrite file and map
*/

type labJob struct {
	name      string
	status    bool
	desc      string
	logger    *log.Entry
	messenger chan slack.MessageInfo
	job
}

type job interface {
	init()
	enable()
	disable()
	commandProcessor(c slack.CommandInfo)
}

type controllerJob struct {
	labJob
	machineName string
	powerStatus bool
	device      any
	customInit  func() (err error)
	customOn    func() (err error)
	customOff   func() (err error)
	logger      *log.Entry
	controller
}

type controller interface {
	init()
	turnOn(ev *slackevents.AppMentionEvent)
	turnOff(ev *slackevents.AppMentionEvent)
	getPowerStatus(ev *slackevents.AppMentionEvent)
	commandProcessor(c slack.CommandInfo)
}

type JobHandler struct {
	jobs   map[string]job
	logger *log.Entry
}

func CreateHandler(m chan slack.MessageInfo) (jh *JobHandler) {
	jobs := make(map[string]job)

	jobLogger := logging.CreateNewLogger("jobhandler", "jobhandler")

	return &JobHandler{
		jobs:   jobs,
		logger: jobLogger,
	}
}

func (jh *JobHandler) InitJobs() {
	for job := range jh.jobs {
		jh.jobs[job].init()
		switch j := jh.jobs[job].(type) {
		case *controllerJob:
			j.customInit()
		}
	}
}

func (jh *JobHandler) CommandReciever(c chan slack.CommandInfo) {
	for command := range c {
		jh.jobs[command.Match].commandProcessor(command)
	}
}

func (lj *labJob) init() {
	lj.status = true
	// lj.messenger <- slack.MessageInfo{
	// 	Text: lj.name + " has been loaded",
	// }
}

func (lj *labJob) enable() {
	lj.status = true
	lj.logger.Info("Enabled job " + lj.name)
}

func (lj *labJob) disable() {
	lj.status = false
	lj.logger.Info("Disabled job " + lj.name)
}

type action func(ev *slackevents.AppMentionEvent)

func (lj *labJob) commandProcessor(c slack.CommandInfo) {}

func (cj *controllerJob) init() {
	cj.labJob.init()
	var message string
	err := cj.customInit()
	if err != nil {
		message = "Couldn't load " + cj.name
		cj.logger.WithField("err", err).Error(message)
	} else {
		message = cj.name + " loaded"
		cj.logger.Info(message)
	}
	cj.messenger <- slack.MessageInfo{
		Text: message,
	}

}

func (cj *controllerJob) turnOn(ev *slackevents.AppMentionEvent) {
	if cj.powerStatus {
		message := "The " + cj.machineName + " is already on"
		go cj.logger.Info(message)
		cj.messenger <- slack.MessageInfo{
			Text:      message,
			ChannelID: ev.Channel,
		}
	} else {
		err := cj.customOn()
		cj.slackPowerResponse(true, err, ev)
	}
}

func (cj *controllerJob) turnOff(ev *slackevents.AppMentionEvent) {
	if !cj.powerStatus {
		message := "The " + cj.machineName + " is already off"
		go cj.logger.Info(message)
		cj.messenger <- slack.MessageInfo{
			Text:      message,
			ChannelID: ev.Channel,
		}
	} else {
		err := cj.customOff()
		cj.slackPowerResponse(false, err, ev)
	}
}

func (cj *controllerJob) slackPowerResponse(status bool, err error, ev *slackevents.AppMentionEvent) {
	statusString := "off"
	if status {
		statusString = "on"
	}
	if err != nil {
		message := "Couldn't turn " + statusString + " the " + cj.machineName
		go cj.logger.WithField("err", err).Error(message)
		cj.messenger <- slack.MessageInfo{
			Text: message,
		}
	} else {
		cj.powerStatus = status
		message := "Turned " + statusString + " the " + cj.machineName
		go cj.logger.Info(message)
		cj.messenger <- slack.MessageInfo{
			Type:      "react",
			Timestamp: ev.TimeStamp,
			ChannelID: ev.Channel,
			Text:      "ok_hand",
		}
		cj.messenger <- slack.MessageInfo{
			Text: message,
		}
	}
}

func (cj *controllerJob) getPowerStatus(ev *slackevents.AppMentionEvent) {
	message := "The " + cj.machineName + " is *off*"
	if cj.powerStatus {
		message = "The " + cj.machineName + " is *on*"
	}
	cj.messenger <- slack.MessageInfo{
		ChannelID: ev.Channel,
		Text:      message,
	}
}

func (cj *controllerJob) commandProcessor(c slack.CommandInfo) {
	if cj.status {
		cropText := strings.ReplaceAll(c.Event.Text, c.Match, "")
		controllerActions := map[string]action{
			"on":     cj.turnOn,
			"off":    cj.turnOff,
			"status": cj.getPowerStatus,
		}
		k := slack.GetKeys(controllerActions)
		match, err := slack.TextMatcher(cropText, k)
		if err == nil {
			f := controllerActions[match]
			f(c.Event)
		} else if err.Error() == "no match found" {
			go cj.logger.WithField("err", err).Warn("No callback function found.")
			cj.messenger <- slack.MessageInfo{
				ChannelID: c.Event.Channel,
				Text:      "I'm not sure what you sayin",
			}
		} else {
			go cj.logger.WithField("err", err).Warn("Many callback functions found.")
			cj.messenger <- slack.MessageInfo{
				ChannelID: c.Event.Channel,
				Text:      "I can respond in multiple ways ...",
			}
		}
	} else {
		cj.messenger <- slack.MessageInfo{
			ChannelID: c.Event.Channel,
			Text:      "The " + cj.name + " is disabled",
		}
	}
}
