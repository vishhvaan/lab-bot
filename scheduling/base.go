package scheduling

import (
	log "github.com/sirupsen/logrus"

	"github.com/go-co-op/gocron"
	"github.com/vishhvaan/lab-bot/logging"
	"github.com/vishhvaan/lab-bot/slack"
)

type scheduleRecord struct {
	ID      string
	Name    string
	CronExp string
	Command slack.CommandInfo
}

type Schedule struct {
	scheduleRecord
	command   slack.CommandInfo
	scheduler *gocron.Scheduler
	logger    *log.Entry
	schedule
}

var schedChan chan *Schedule

// // type SchedJobs struct {
// // 	name      string
// // 	status    string
// // 	desc      string
// // 	frequency string
// // 	sched     *gocron.Scheduler
// // 	logger    *log.Entry
// // }

// // type SchedBirthdays struct {
// // 	enabled bool

// // }

type schedule interface {
	init()
	enable()
	disable()
	commandProcessor(c slack.CommandInfo)
}

type ScheduleTracker struct {
	schedules map[string]*Schedule
	messenger chan slack.MessageInfo
	logger    *log.Entry
}

func CreateScheduleTracker() (st *ScheduleTracker) {
	schedules := make(map[string]*Schedule)

	schedLogger := logging.CreateNewLogger("scheduling", "scheduling")

	return &ScheduleTracker{
		schedules: schedules,
		logger:    schedLogger,
	}
}

// func (jh *SchedHandler) InitScheds() {
// 	for job := range jh.jobs {
// 		jh.jobs[job].init()
// 		switch j := jh.jobs[job].(type) {
// 		case *controllerJob:
// 			j.customInit()
// 		}
// 	}
// }

func (st *ScheduleTracker) Reciever() {
	for sched := range schedChan {
		if sched.scheduler.IsRunning() {
			st.schedules[sched.ID] = sched
		} else {
			delete(st.schedules, sched.ID)
		}
	}
}
