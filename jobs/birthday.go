package jobs

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"

	"github.com/vishhvaan/lab-bot/db"
	"github.com/vishhvaan/lab-bot/functions"
	"github.com/vishhvaan/lab-bot/scheduling"
	"github.com/vishhvaan/lab-bot/slack"
)

type birthdayJob struct {
	labJob
	dbPath         []string
	birthdayDbPath []string
	scheduling     scheduling.BirthdaySchedule
}

func (bj *birthdayJob) init() {
	bj.labJob.init()

	bj.dbPath = append([]string{"jobs", "controller"}, bj.keyword)
	bj.birthdayDbPath = append(bj.dbPath, bj.keyword)

	// ensure database is there or create database
	bj.scheduling.Sched = make(map[string]*scheduling.Schedule)
	bj.scheduling.DbPath = append(bj.dbPath, "scheduling")
	bj.checkCreateBucket()
	numBirthdays, err := bj.numerateBirthdays()

	if err == nil {
		bj.logger.Info(bj.name + " loaded")
		slack.Message(bj.name + " loaded. " + strconv.Itoa(numBirthdays) + "birthdays found.")
	} else {
		bj.logger.WithError(err).Error(bj.name + " cannot initialize")
	}

}

func (bj *birthdayJob) commandProcessor(c slack.CommandInfo) {
	if bj.active {
		birthdayActions := map[string]action{
			"record":   bj.recordBirthday,
			"status":   bj.birthdayStatus,
			"upcoming": bj.upcomingBirthdays,
		}
		if len(c.Fields) == 1 {
			bj.birthdayStatus(c)
		} else {
			k := functions.GetKeys(birthdayActions)
			subcommand := strings.ToLower(c.Fields[1])
			if functions.Contains(k, subcommand) {
				f := birthdayActions[subcommand]
				f(c)
			} else {
				bj.errorMsg(c.Fields, c.Channel, "Wrong syntax, young padwan")
			}
		}
	} else {
		slack.PostMessage(c.Channel, "The "+bj.name+" is disabled")
	}
}

// db organization birthdays/key = user, value = time.Time

func (bj *birthdayJob) checkCreateBucket() (exists bool) {
	if !db.CheckBucketExists(bj.birthdayDbPath) {
		db.CreateBucket(bj.birthdayDbPath)
	}

	if !db.CheckBucketExists(bj.scheduling.DbPath) {
		db.CreateBucket(bj.scheduling.DbPath)
	}

	return exists
}

func (bj *birthdayJob) errorMsg(c slack.CommandInfo, err error, message string) {
	go bj.logger.WithField("fields", c.Fields).WithError(err).Warn(message)
	slack.PostMessage(c.Channel, message)
}

func (bj *birthdayJob) numerateBirthdays() (numBirthdays int, err error) {
	keys, _, err := db.GetAllKeysValues(bj.birthdayDbPath)
	return len(keys), nil
}

// birthday record 10-24-2000
func (bj *birthdayJob) recordBirthday(c slack.CommandInfo) {
	if len(c.Fields) >= 3 {
		newBirthday, err := dateparse.ParseAny(c.Fields[3])
		if err != nil {
			bj.errorMsg(c, err, "cannot parse date")
			return
		}

		b, err := db.ReadValue(bj.birthdayDbPath, c.User)
		if err != nil {
			bj.errorMsg(c, err, "cannot read existing birthday from db")
			return
		}

		if b == nil {
			b, err := json.Marshal(newBirthday)
			if err != nil {
				bj.errorMsg(c, err, "cannot convert birthday into json")
				return
			}
			err = db.AddValue(bj.birthdayDbPath, c.User, b)
			if err != nil {
				bj.errorMsg(c, err, "cannot record birthday to database")
			}
		} else {
			var oldBirthday time.Time
			err = json.Unmarshal(b, oldBirthday)
			if err != nil {
				bj.errorMsg(c, err, "cannot read existing birthday from db")
				return
			}

			if oldBirthday == newBirthday {
				slack.SendMessage(c.Channel, "This birthday is already on record")
			} else {
				slack.SendMessage(c.Channel, "There is a different birthday already on record for you, please delete it before entering a new one")
			}
		}
	}
}

func (bj *birthdayJob) birthdayStatus(c slack.CommandInfo) {
	b, err := db.ReadValue(bj.birthdayDbPath, c.User)
	if err != nil {
		bj.errorMsg(c, err, "cannot read existing birthday from db")
		return
	}

	if b == nil {
		slack.SendMessage(c.Channel, "No birthday in the database. Use the command 'birthday record 2006-10-24' to record your birthday")
	} else {
		var birthday time.Time
		json.Unmarshal(b, birthday)
		slack.SendMessage(c.Channel, "Your birthday is on "+birthday.Format(time.DateOnly))
	}
}
