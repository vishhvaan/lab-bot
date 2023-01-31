package jobs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vishhvaan/lab-bot/functions"
	"github.com/vishhvaan/lab-bot/slack"
)

type labMeetingJob struct {
	labJob
	labMeetingGroups map[string][]string
}

func (lm *labMeetingJob) init() {
	lm.labJob.init()

	lm.labMeetingGroups = make(map[string][]string)
	// lm.loadlabMeetingGroupsFromDB()
}

func (lm *labMeetingJob) commandProcessor(c slack.CommandInfo) {
	if lm.active {
		controllerActions := map[string]action{
			"groups":  lm.groupsHandler,
			"present": lm.presentHandler,
		}
		if len(c.Fields) == 1 {
			lm.printlabMeetingGroups(c)
		} else {
			k := functions.GetKeys(controllerActions)
			subcommand := strings.ToLower(c.Fields[1])
			if functions.Contains(k, subcommand) {
				f := controllerActions[subcommand]
				f(c)
			} else {
				lm.errorMsg(c.Fields, c.Channel, "I'm not sure what you sayin")
			}
		}
	} else {
		slack.PostMessage(c.Channel, "The "+lm.name+" is disabled")
	}
}

func (lm *labMeetingJob) groupsHandler(c slack.CommandInfo) {
	controllerActions := map[string]action{
		"json":   lm.printlabMeetingGroupsJSON,
		"update": lm.parselabMeetingGroups,
	}
	if len(c.Fields) == 2 {
		lm.printlabMeetingGroups(c)
	} else {
		k := functions.GetKeys(controllerActions)
		subcommand := strings.ToLower(c.Fields[2])
		if functions.Contains(k, subcommand) {
			f := controllerActions[subcommand]
			f(c)
		} else {
			lm.errorMsg(c.Fields, c.Channel, "I'm not sure what you sayin")
		}
	}
}

func (lm *labMeetingJob) presentHandler(c slack.CommandInfo) {

}

func (lm *labMeetingJob) parselabMeetingGroups(c slack.CommandInfo) {
	if len(c.Fields) >= 3 {
		groupsJSON := strings.Join(c.Fields[4:], " ")
		err := json.Unmarshal([]byte(groupsJSON), &lm.labMeetingGroups)
		if err != nil {
			go lm.logger.WithField("command", groupsJSON).WithError(err).Warn("Cannot unmarshal json from message")
			slack.PostMessage(c.Channel, "Cannot parse groups from the input JSON")
			return
		}
	}
	lm.errorMsg(c.Fields, c.Channel, "Malformed groups update command")
}

// func (lm *LabMeetingJob) loadlabMeetingGroupsFromDB() {

// }

func (lm *labMeetingJob) printlabMeetingGroups(c slack.CommandInfo) {
	if lm.labMeetingGroups != nil && len(lm.labMeetingGroups) == 0 {
		lm.sendMsg(c.Channel, "Lab Meeting Groups: "+fmt.Sprint(lm.labMeetingGroups))
	} else {
		lm.errorMsg(c.Fields, c.Channel, "Lab meeting groups are not defined")
	}
}

func (lm *labMeetingJob) printlabMeetingGroupsJSON(c slack.CommandInfo) {
	if lm.labMeetingGroups != nil && len(lm.labMeetingGroups) == 0 {
		str, err := json.Marshal(lm.labMeetingGroups)
		if err != nil {
			lm.errorMsg(c.Fields, c.Channel, "Cannot parse internal groups into json")
		} else {
			lm.sendMsg(c.Channel, "Lab Meeting Groups: "+string(str))
		}
	} else {
		lm.errorMsg(c.Fields, c.Channel, "Lab meeting groups are not defined")
	}
}

func (lm *labMeetingJob) sendMsg(channel string, message string) {
	go lm.logger.Info(message)
	slack.PostMessage(channel, message)
}

func (lm *labMeetingJob) errorMsg(fields []string, channel string, message string) {
	go lm.logger.WithField("fields", fields).Warn(message)
	slack.PostMessage(channel, message)
}
