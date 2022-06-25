package jobs

import (
	// "github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/pkg/logging"
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
	name   string
	status bool
	desc   string
	logger *log.Entry
	job
}

type job interface {
	init()
}

type controllerJob struct {
	labJob
	powerStatus bool
	controller
	logger *log.Entry
}

type controller interface {
	turnOn()
	turnOff()
}

type jobHandler struct {
	logger *log.Entry
}

func CreateHandler() (jh *jobHandler) {
	jobLogger := logging.CreateNewLogger("jobhandler", "jobhandler")

	return &jobHandler{
		logger: jobLogger,
	}
}
