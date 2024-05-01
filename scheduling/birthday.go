package scheduling

import (
	log "github.com/sirupsen/logrus"
)

type BirthdaySchedule struct {
	birthdayMessageChannel string
	Logger                 *log.Entry
	Sched                  map[string]*Schedule
	DbPath                 []string
}
