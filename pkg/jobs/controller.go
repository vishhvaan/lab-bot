package jobs

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack/slackevents"

	"github.com/vishhvaan/lab-bot/pkg/functions"
	"github.com/vishhvaan/lab-bot/pkg/scheduling"
	"github.com/vishhvaan/lab-bot/pkg/slack"
)

type controllerJob struct {
	labJob
	machineName string
	powerStatus bool
	lastPowerOn time.Time
	device      any
	customInit  func() (err error)
	customOn    func() (err error)
	customOff   func() (err error)
	logger      *log.Entry
	scheduling  scheduling.ControllerSchedule
	controller
}

type controller interface {
	init()
	turnOn(channel string)
	turnOff(channel string)
	getPowerStatus(channel string)
	commandProcessor(c slack.CommandInfo)
}

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

func (cj *controllerJob) turnOn(c slack.CommandInfo) {
	if commandCheck(c, 2, cj.messenger, cj.logger) {
		if cj.powerStatus {
			message := "The " + cj.machineName + " is already on"
			go cj.logger.Info(message)
			cj.messenger <- slack.MessageInfo{
				Text:      message,
				ChannelID: c.Event.Channel,
			}
		} else {
			err := cj.customOn()
			cj.lastPowerOn = time.Now()
			cj.slackPowerResponse(true, err, c.Event)
		}
	}
}

func (cj *controllerJob) turnOff(c slack.CommandInfo) {
	if commandCheck(c, 2, cj.messenger, cj.logger) {
		if !cj.powerStatus {
			message := "The " + cj.machineName + " is already off"
			go cj.logger.Info(message)
			cj.messenger <- slack.MessageInfo{
				Text:      message,
				ChannelID: c.Event.Channel,
			}
		} else {
			err := cj.customOff()
			cj.slackPowerResponse(false, err, c.Event)
		}
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

func (cj *controllerJob) getPowerStatus(c slack.CommandInfo) {
	if commandCheck(c, 2, cj.messenger, cj.logger) {
		message := "The " + cj.machineName + " is "
		if cj.powerStatus {
			uptime := time.Since(cj.lastPowerOn).Round(time.Second)
			message += "*on*\nUptime: " + fmt.Sprint(uptime)
		} else {
			message += "*off*"
		}
		cj.messenger <- slack.MessageInfo{
			ChannelID: c.Event.Channel,
			Text:      message,
		}
	}
}

func (cj *controllerJob) commandProcessor(c slack.CommandInfo) {
	if cj.status {
		controllerActions := map[string]action{
			"on":       cj.turnOn,
			"off":      cj.turnOff,
			"status":   cj.getPowerStatus,
			"schedule": cj.scheduleHandler,
		}
		if len(c.Fields) == 1 {
			cj.getPowerStatus(c)
		} else {
			k := functions.GetKeys(controllerActions)
			subcommand := strings.ToLower(c.Fields[1])
			if functions.Contains(k, subcommand) {
				f := controllerActions[subcommand]
				if subcommand != "schedule" {
					if commandCheck(c, 2, cj.messenger, cj.logger) {
						f(c)
					}
				} else {
					f(c)
				}
			} else {
				go cj.logger.WithField("fields", c.Fields).Warn("No callback function found.")
				cj.messenger <- slack.MessageInfo{
					ChannelID: c.Event.Channel,
					Text:      "I'm not sure what you sayin",
				}
			}
		}
	} else {
		cj.messenger <- slack.MessageInfo{
			ChannelID: c.Event.Channel,
			Text:      "The " + cj.name + " is disabled",
		}
	}
}

func (cj *controllerJob) scheduleHandler(c slack.CommandInfo) {
	schedulingActions := map[string]action{
		"set":    cj.setSched,
		"remove": cj.removeSched,
		"status": cj.sendSchedulingStatus,
	}
	if len(c.Fields) == 2 {
		cj.sendSchedulingStatus(c)
	} else if len(c.Fields) > 2 {
		k := functions.GetKeys(schedulingActions)
		subcommand := strings.ToLower(c.Fields[1])
		if functions.Contains(k, subcommand) {
			f := schedulingActions[subcommand]
			f(c)
		} else {
		}
	}
}

func (cj *controllerJob) sendSchedulingStatus(c slack.CommandInfo) {

}