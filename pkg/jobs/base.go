package jobs

import (
	// "github.com/go-co-op/gocron"
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
	commandProcessor(ev *slackevents.AppMentionEvent)
}

type controllerJob struct {
	labJob
	powerStatus bool
	controller
	logger *log.Entry
}

type controller interface {
	init()
	turnOn()
	turnOff()
	commandProcessor(ev *slackevents.AppMentionEvent)
}

type JobHandler struct {
	jobs   map[string]job
	logger *log.Entry
}

func CreateHandler(m chan slack.MessageInfo) (jh *JobHandler) {
	var jobs map[string]job
	jobLogger := logging.CreateNewLogger("jobhandler", "jobhandler")

	return &JobHandler{
		jobs:   jobs,
		logger: jobLogger,
	}
}

func (jh *JobHandler) CommandReciever(c chan slack.CommandInfo) {
	for command := range c {
		jh.jobs[command.Match].commandProcessor(command.Event)
	}
}

func (lj *labJob) init() {
	lj.status = true
	lj.messenger <- slack.MessageInfo{
		Text: lj.name + " has been loaded",
	}
}

func (lj *labJob) enable() {
	lj.status = true
}

func (lj *labJob) disable() {
	lj.status = false
}

func (cj *controllerJob) init() {
	cj.labJob.init()
}
