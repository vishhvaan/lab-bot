package jobs

import (
	"strings"

	"github.com/vishhvaan/lab-bot/functions"
	"github.com/vishhvaan/lab-bot/scheduling"
	"github.com/vishhvaan/lab-bot/slack"
)

type birthdayJob struct {
	labJob
	dbPath     []string
	scheduling scheduling.BirthdaySchedule
}

func (bj *birthdayJob) init() {
	bj.labJob.init()

	bj.dbPath = append([]string{"jobs", "controller"}, bj.keyword)

	// ensure database is there or create database

}

func (bj *birthdayJob) commandProcessor(c slack.CommandInfo) {
	if bj.active {
		birthdayActions := map[string]action{
			"record":   bj.recordBirthday,
			"status":   bj.birthdayStatus,
			"upcoming": bj.upcomingBirthdays,
		}
		if len(c.Fields) == 1 {
			bj.birthdayStatus(c)
		} else {
			k := functions.GetKeys(birthdayActions)
			subcommand := strings.ToLower(c.Fields[1])
			if functions.Contains(k, subcommand) {
				f := birthdayActions[subcommand]
				f(c)
			} else {
				bj.errorMsg(c.Fields, c.Channel, "Wrong syntax, young padwan")
			}
		}
	} else {
		slack.PostMessage(c.Channel, "The "+bj.name+" is disabled")
	}
}

func (bj *birthdayJob) errorMsg(fields []string, channel string, message string) {
	go bj.logger.WithField("fields", fields).Warn(message)
	slack.PostMessage(channel, message)
}
