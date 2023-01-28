package jobs

import (
	"strings"

	"github.com/vishhvaan/lab-bot/functions"
	"github.com/vishhvaan/lab-bot/slack"
)

type LabMeetingJob struct {
	labJob
	labMeetingGroups map[string][]string
}

func (lm *LabMeetingJob) init() {
	lm.labJob.init()

}

func (lm *LabMeetingJob) commandProcessor(c slack.CommandInfo) {
	if lm.active {
		controllerActions := map[string]action{
			"groups":  lm.groupsHandler,
			"present": lm.presentHandler,
		}
		if len(c.Fields) == 1 {
			lm.printlabMeetingGroups()
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

func (lm *LabMeetingJob) groupsHandler(c slack.CommandInfo) {

}

func (lm *LabMeetingJob) presentHandler(c slack.CommandInfo) {

}

func (lm *LabMeetingJob) parselabMeetingGroups(url string) {

}

func (lm *LabMeetingJob) printlabMeetingGroups() {

}

func (lm *LabMeetingJob) errorMsg(fields []string, channel string, message string) {
	go lm.logger.WithField("fields", fields).Warn(message)
	slack.PostMessage(channel, message)
}
