package scheduling

import (
	"encoding/json"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/db"
	"github.com/vishhvaan/lab-bot/slack"
)

type BirthdaySchedule struct {
	birthdayMessageChannel string
	dbPath                 []string
	Logger                 *log.Entry
	sched                  map[string]*Schedule
}

func (bs *BirthdaySchedule) Init(dbPath []string, logger *log.Entry) {
	bs.sched = make(map[string]*Schedule)
	bs.dbPath = dbPath
}

func (bs *BirthdaySchedule) UpcomingBirthdays(c slack.CommandInfo) {
	lU, err := db.ReadValue(append(bs.dbPath, "upcoming"), "lastUpdated")
	if err != nil {
		bs.errorMsg(c, err, "cannot decode birthday")
		return
	}

	var lastUpdated time.Time
	var upcomingBirthdays map[string][]string

	if lU == nil {
		upcomingBirthdays, err = bs.generateUpcomingBirthdays()
		if err != nil {
			bs.errorMsg(c, err, "cannot generate upcoming birthdays")
			return
		}
	} else {
		err = lastUpdated.UnmarshalJSON(lU)
		if err != nil {
			bs.errorMsg(c, err, "cannot decode birthday")
			return
		}

		if lastUpdated.Truncate(24*time.Hour) == time.Now().Truncate(24*time.Hour) {
			uB, err := db.ReadValue(append(bs.dbPath, "upcoming"), "birthdays")
			if err != nil {
				bs.errorMsg(c, err, "cannot decode birthday")
				return
			}

			err = json.Unmarshal(uB, &upcomingBirthdays)
			if err != nil {
				bs.errorMsg(c, err, "cannot read upcoming birthdays from db")
				return
			}
		} else {
			upcomingBirthdays, err = bs.generateUpcomingBirthdays()
			if err != nil {
				bs.errorMsg(c, err, "cannot generate upcoming birthdays")
				return
			}
		}
	}

	message, err := bs.formatUpcomingBirthdays(upcomingBirthdays)

	if err != nil {
		bs.errorMsg(c, err, "cannot format upcoming birthdays")
	}
	slack.PostMessage(c.Channel, message)

}

func (bs *BirthdaySchedule) generateUpcomingBirthdays() (upcomingBirthdays map[string][]string, err error) {
	upcomingBirthdays = make(map[string][]string)

	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	tomorrow := today.AddDate(0, 0, 1)
	nextWeek := today.AddDate(0, 0, 7)
	nextMonth := today.AddDate(0, 1, 0)

	updateUpcoming := func(userKey []byte, birthdayValue []byte) error {
		var birthday time.Time
		err := birthday.UnmarshalJSON(birthdayValue)
		if err != nil {
			return err
		}
		birthday = time.Date(time.Now().Year(), birthday.Month(), birthday.Day(), 0, 0, 0, 0, birthday.Location())
		if birthday == today {
			upcomingBirthdays["todayBDs"] = append(upcomingBirthdays["todayBDs"], string(userKey))
		} else if birthday == tomorrow {
			upcomingBirthdays["nextDayBDs"] = append(upcomingBirthdays["nextDayBDs"], string(userKey))
		} else if birthday.Before(nextWeek) {
			upcomingBirthdays["nextWeekBDs"] = append(upcomingBirthdays["nextWeekBDs"], string(userKey))
		} else if birthday.Before(nextMonth) {
			upcomingBirthdays["nextMonthBDs"] = append(upcomingBirthdays["nextMonthBDs"], string(userKey))
		}
		return nil
	}

	err = db.RunCallbackOnEachKey(append(bs.dbPath, "records"), updateUpcoming)
	if err != nil {
		return upcomingBirthdays, err
	}
	n, err := now.MarshalJSON()
	if err != nil {
		return upcomingBirthdays, err
	}
	b, err := json.Marshal(upcomingBirthdays)
	if err != nil {
		return upcomingBirthdays, err
	}
	err = db.AddValue(append(bs.dbPath, "upcoming"), "birthdays", b)
	if err != nil {
		return upcomingBirthdays, err
	}
	err = db.AddValue(append(bs.dbPath, "upcoming"), "lastUpdated", n)
	if err != nil {
		return upcomingBirthdays, err
	}

	return upcomingBirthdays, nil
}

func (bs *BirthdaySchedule) formatUpcomingBirthdays(upcomingBirthdays map[string][]string) (message string, err error) {
	var m strings.Builder

	genUsers := func(users []string) {
		if len(users) == 0 {
			m.WriteString("none\n")
		} else {
			for i, u := range users {
				m.WriteString(slack.GetUserName(u))
				if i+1 < len(users) {
					m.WriteString(", ")
				}
			}
			m.WriteString("\n")
		}
	}

	m.WriteString("Upcoming Birthdays:\n")
	m.WriteString("Today: ")
	genUsers(upcomingBirthdays["todayBDs"])
	m.WriteString("Tomorrow: ")
	genUsers(upcomingBirthdays["nextDayBDs"])
	m.WriteString("This week: ")
	genUsers(upcomingBirthdays["nextWeekBDs"])
	m.WriteString("This month: ")
	genUsers(upcomingBirthdays["nextMonthBDs"])

	return m.String(), nil
}

func (bs *BirthdaySchedule) errorMsg(c slack.CommandInfo, err error, message string) {
	go bs.Logger.WithField("fields", c.Fields).WithError(err).Warn(message)
	slack.PostMessage(c.Channel, message)
}
