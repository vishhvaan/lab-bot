package jobs

import (
	"github.com/go-co-op/gocron"
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
	name      string
	status    string
	desc      string
	frequency string
	sched     *gocron.Scheduler
	logger    *log.Entry
}

type jobHandler struct {
	techRemind *labJob
	logger     *log.Entry
}

func (lj labJob) initJob(name string, logger *log.Entry) {
	lj.name = name
	lj.logger = logger.WithField("job", name)
}

func CreateHandler() (jh *jobHandler) {
	jobLogger := logging.CreateNewLogger("jobhandler", "jobhandler")

	var tr labJob
	tr.initJob("techremind", jobLogger)

	return &jobHandler{
		techRemind: &tr,
		logger:     jobLogger,
	}
}
