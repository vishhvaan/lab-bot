package scheduling

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/db"
	"github.com/vishhvaan/lab-bot/slack"
)

type BirthdaySchedule struct {
	birthdayMessageChannel string
	dbPath                 []string
	logger                 *log.Entry
	sched                  map[string]*Schedule
}

func (bs *BirthdaySchedule) Init(dbPath []string) {
	bs.sched = make(map[string]*Schedule)
	bs.dbPath = dbPath
}

func (bs *BirthdaySchedule) UpcomingBirthdays(c slack.CommandInfo) {

}

func (bs *BirthdaySchedule) generateUpcomingBirthdays(c slack.CommandInfo) map[string][]string {
	upcomingBirthdays := make(map[string][]string)

	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	tomorrow := today.AddDate(0, 0, 1)
	nextWeek := today.AddDate(0, 0, 7)
	nextMonth := today.AddDate(0, 1, 0)

	updateUpcoming := func(userKey []byte, birthdayValue []byte) {
		var birthday time.Time
		err := birthday.UnmarshalJSON(birthdayValue)
		if err != nil {
			bs.errorMsg(c, err, "cannot decode birthday")
			return
		}
		birthday = birthday.Truncate(24 * time.Hour)
		if birthday == today {
			upcomingBirthdays["todayBDs"] = append(upcomingBirthdays["todayBDs"], string(userKey))
		} else if birthday == tomorrow {
			upcomingBirthdays["nextDayBDs"] = append(upcomingBirthdays["nextDayBDs"], string(userKey))
		} else if birthday.Before(nextWeek) {
			upcomingBirthdays["nextWeekBDs"] = append(upcomingBirthdays["nextWeekBDs"], string(userKey))
		} else if birthday.Before(nextMonth) {
			upcomingBirthdays["nextMonthBDs"] = append(upcomingBirthdays["nextMonthBDs"], string(userKey))
		}
	}

	db.RunCallbackOnEachKey(append(bs.dbPath, "records"), updateUpcoming)
	n, err := now.MarshalJSON()
	if err != nil {
		bs.errorMsg(c, err, "convert time to json")
	}
	b, err := json.Marshal(upcomingBirthdays)
	if err != nil {
		bs.errorMsg(c, err, "convert birthday map to json")
	}
	db.AddValue(append(bs.dbPath, "upcoming"), "birthdays", b)
	db.AddValue(append(bs.dbPath, "upcoming"), "lastUpdated", n)
	return upcomingBirthdays
}

func (bs *BirthdaySchedule) errorMsg(c slack.CommandInfo, err error, message string) {
	go bs.logger.WithField("fields", c.Fields).WithError(err).Warn(message)
	slack.PostMessage(c.Channel, message)
}
