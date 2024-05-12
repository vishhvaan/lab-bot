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
	var force bool
	if len(c.Fields) > 2 {
		if c.Fields[2] == "force" {
			force = true
		} else {
			slack.SendMessage(c.Channel, "only the force flag is supported")
			return
		}
	}

	lU, err := db.ReadValue(append(bs.dbPath, "upcoming"), "lastUpdated")
	if err != nil {
		bs.errorMsg(c, err, "cannot decode birthday")
		return
	}

	var lastUpdated time.Time
	var upcomingBirthdays map[string]map[string]time.Time

	if lU == nil {
		upcomingBirthdays, err = bs.generateUpcomingBirthdays()
		if err != nil {
			bs.errorMsg(c, err, "cannot generate upcoming birthdays")
			return
		}
	} else {
		err = lastUpdated.UnmarshalJSON(lU)
		if err != nil {
			bs.errorMsg(c, err, "cannot decode time birthday message was last updated")
			return
		}

		if lastUpdated.Truncate(24*time.Hour) == time.Now().Truncate(24*time.Hour) && !force {
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

func (bs *BirthdaySchedule) generateUpcomingBirthdays() (upcomingBirthdays map[string]map[string]time.Time, err error) {
	upcomingBirthdays = make(map[string]map[string]time.Time)
	upcomingBirthdays["todayBDs"] = make(map[string]time.Time)
	upcomingBirthdays["nextDayBDs"] = make(map[string]time.Time)
	upcomingBirthdays["nextWeekBDs"] = make(map[string]time.Time)
	upcomingBirthdays["nextMonthBDs"] = make(map[string]time.Time)

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

		birthday = time.Date(time.Now().Year(), birthday.Month(), birthday.Day(), birthday.Hour(), birthday.Minute(), birthday.Second(), birthday.Nanosecond(), birthday.Location())

		if birthday == today {
			upcomingBirthdays["todayBDs"][string(userKey)] = birthday
		} else if birthday == tomorrow {
			upcomingBirthdays["nextDayBDs"][string(userKey)] = birthday
		} else if birthday.Before(nextWeek) {
			upcomingBirthdays["nextWeekBDs"][string(userKey)] = birthday
		} else if birthday.Before(nextMonth) {
			upcomingBirthdays["nextMonthBDs"][string(userKey)] = birthday
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

func (bs *BirthdaySchedule) formatUpcomingBirthdays(upcomingBirthdays map[string]map[string]time.Time) (message string, err error) {
	var m strings.Builder

	genUsers := func(users map[string]time.Time) (s string) {
		var v strings.Builder
		if len(users) == 0 {
			return "none\n"
		} else {
			for user, birthday := range users {
				v.WriteString(slack.GetUserName(user) + " [" + birthday.UTC().Format("Jan 02") + "], ")
			}
			o := v.String()
			return o[:len(o)-2] + "\n"
		}
	}

	m.WriteString("Upcoming Birthdays:\n")
	m.WriteString("Today: ")
	m.WriteString(genUsers(upcomingBirthdays["todayBDs"]))
	m.WriteString("Tomorrow: ")
	m.WriteString(genUsers(upcomingBirthdays["nextDayBDs"]))
	m.WriteString("Next 7 Days: ")
	m.WriteString(genUsers(upcomingBirthdays["nextWeekBDs"]))
	m.WriteString("Next 30 Days: ")
	m.WriteString(genUsers(upcomingBirthdays["nextMonthBDs"]))

	return m.String(), nil
}

func (bs *BirthdaySchedule) errorMsg(c slack.CommandInfo, err error, message string) {
	go bs.Logger.WithField("fields", c.Fields).WithError(err).Warn(message)
	slack.PostMessage(c.Channel, message)
}
