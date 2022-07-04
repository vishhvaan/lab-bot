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
	keyword   string
	responses map[string]action
	job
}

type action func(ev *slackevents.AppMentionEvent)

type job interface {
	init()
	enable()
	disable()
	commandProcessor(c slack.CommandInfo)
}

type JobHandler struct {
	jobs   map[string]job
	messenger chan slack.MessageInfo
	logger *log.Entry
}

func CreateHandler(m chan slack.MessageInfo) (jh *JobHandler) {
	jobs := make(map[string]job)

	jobLogger := logging.CreateNewLogger("jobhandler", "jobhandler")

	return &JobHandler{
		jobs:   jobs,
		messenger: m,
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
		k := strings.ToLower(fields[0])
		for job := range jh.jobs {
			if jh.jobs[job].keyword == k {
				jh.jobs[job].commandProcessor(command.Event)
				return
			}
		}
		jh.messenger <- slack.MessageInfo{
			Text: "Couldn't find a response to your command.",
			ChannelID: c.Event.Text,
		}
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

func (lj *labJob) commandProcessor(c slack.CommandInfo) {}
