package jobs

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/functions"
	"github.com/vishhvaan/lab-bot/logging"
	"github.com/vishhvaan/lab-bot/slack"
)

type labJob struct {
	name      string
	keyword   string
	active    bool
	desc      string
	logger    *log.Entry
	responses map[string]action
	job
}

type action func(c slack.CommandInfo)

type job interface {
	init()
	enable()
	disable()
	commandProcessor(c slack.CommandInfo)
}

type JobHandler struct {
	jobs   map[string]job
	logger *log.Entry
}

func CreateHandler() (jh *JobHandler) {
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

func (jh *JobHandler) CommandReceiver() {
	for command := range slack.CommandChan {
		k := strings.ToLower(command.Fields[0])
		if functions.Contains(functions.GetKeys(jh.jobs), k) {
			jh.jobs[k].commandProcessor(command)
		} else {
			slack.MessageChan <- slack.MessageInfo{
				Text:      "I couldn't find a response to your command.",
				ChannelID: command.Channel,
			}
		}
	}
}

func (lj *labJob) init() {
	lj.active = true
	// lj.messenger <- slack.MessageInfo{
	// 	Text: lj.name + " has been loaded",
	// }
}

func (lj *labJob) enable() {
	lj.active = true
	lj.logger.Info("Enabled job " + lj.name)
}

func (lj *labJob) disable() {
	lj.active = false
	lj.logger.Info("Disabled job " + lj.name)
}

func (lj *labJob) commandProcessor(c slack.CommandInfo) {}

func commandCheck(c slack.CommandInfo, length int, m chan slack.MessageInfo, l *log.Entry) bool {
	if len(c.Fields) > length {
		message := "Your command has more parameters than necessary"
		go l.Info(message)
		m <- slack.MessageInfo{
			Text:      message,
			ChannelID: c.Channel,
		}
		return false
	} else {
		return true
	}
}
