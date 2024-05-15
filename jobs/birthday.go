package jobs

import (
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
	dbPath     []string
	scheduling scheduling.BirthdaySchedule
}

func (bj *birthdayJob) init() {
	bj.labJob.init()

	bj.dbPath = append([]string{"jobs", "controller"}, bj.keyword)

	// ensure database is there or create database
	bj.scheduling.Init(bj.dbPath, bj.logger)

	bj.checkCreateBucket()
	numBirthdays, err := bj.numerateBirthdays()

	if err == nil {
		bj.logger.Info(bj.name + " loaded")
		slack.Message(bj.name + " loaded. " + strconv.Itoa(numBirthdays) + " birthdays found.")
	} else {
		bj.logger.WithError(err).Error(bj.name + " cannot initialize")
	}

}

func (bj *birthdayJob) commandProcessor(c slack.CommandInfo) {
	if bj.active {
		birthdayActions := map[string]action{
			"record":   bj.recordBirthday,
			"delete":   bj.deleteBirthday,
			"status":   bj.birthdayStatus,
			"upcoming": bj.scheduling.UpcomingBirthdays,
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
				slack.PostMessage(c.Channel, "Wrong syntax, young padwan")
				bj.logger.WithField("fields", c.Fields).Info("Wrong syntax for birthday")
			}
		}
	} else {
		slack.PostMessage(c.Channel, "The "+bj.name+" is disabled")
	}
}

// db organization birthdays/key = user, value = time.Time

func (bj *birthdayJob) checkCreateBucket() {
	if !db.CheckBucketExists(append(bj.dbPath, "records")) {
		db.CreateBucket(append(bj.dbPath, "records"))
	}

	if !db.CheckBucketExists(append(bj.dbPath, "schedule")) {
		db.CreateBucket(append(bj.dbPath, "schedule"))
	}

	if !db.CheckBucketExists(append(bj.dbPath, "upcoming")) {
		db.CreateBucket(append(bj.dbPath, "upcoming"))
	}
}

func (bj *birthdayJob) errorMsg(c slack.CommandInfo, err error, message string) {
	go bj.logger.WithField("fields", c.Fields).WithError(err).Warn(message)
	slack.PostMessage(c.Channel, message)
}

func (bj *birthdayJob) numerateBirthdays() (numBirthdays int, err error) {
	keys, _, err := db.GetAllKeysValues(append(bj.dbPath, "records"))
	return len(keys), err
}

func (bj *birthdayJob) birthdayStatus(c slack.CommandInfo) {
	b, err := db.ReadValue(append(bj.dbPath, "records"), c.User)
	if err != nil {
		bj.errorMsg(c, err, "cannot read existing birthday from db")
		return
	}

	if b == nil {
		slack.SendMessage(c.Channel, "No birthday in the database. Use the command 'birthday record 2006-10-24' to record your birthday")
	} else {
		var birthday time.Time
		birthday.UnmarshalJSON(b)
		slack.SendMessage(c.Channel, "Your birthday is on "+birthday.UTC().Format("Jan 02"))
	}
}

// birthday record 10-24-2000
func (bj *birthdayJob) recordBirthday(c slack.CommandInfo) {
	var force bool
	if len(c.Fields) > 3 {
		if c.Fields[3] == "force" {
			force = true
		} else {
			slack.SendMessage(c.Channel, "only the force flag is supported")
			return
		}
	}

	if len(c.Fields) >= 3 {
		newBirthday, err := dateparse.ParseAny(c.Fields[2])
		if err != nil {
			go bj.logger.WithField("fields", c.Fields).WithError(err).Warn("cannot parse date")
			slack.PostMessage(c.Channel, "cannot parse date: "+err.Error())
			return
		}
		newBirthday = newBirthday.Truncate(24 * time.Hour).In(time.Now().Location())

		b, err := db.ReadValue(append(bj.dbPath, "records"), c.User)
		if err != nil {
			bj.errorMsg(c, err, "cannot read existing birthday from db")
			return
		}

		if b == nil || force {
			b, err := newBirthday.MarshalJSON()
			if err != nil {
				bj.errorMsg(c, err, "cannot convert birthday into json")
				return
			}
			err = db.AddValue(append(bj.dbPath, "records"), c.User, b)
			if err != nil {
				bj.errorMsg(c, err, "cannot record birthday to database")
			}

			slack.React(c.TimeStamp, c.Channel, "tada")
		} else {
			var oldBirthday time.Time
			err = oldBirthday.UnmarshalJSON(b)
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

func (bj *birthdayJob) deleteBirthday(c slack.CommandInfo) {
	if len(c.Fields) > 2 {
		go bj.logger.WithField("fields", c.Fields).Warn("too many fields")
		slack.SendMessage(c.Channel, "This birthday is already on record")
		return
	}

	b, err := db.ReadValue(append(bj.dbPath, "records"), c.User)
	if err != nil {
		bj.errorMsg(c, err, "cannot read existing birthday from db")
		return
	}

	if b == nil {
		slack.SendMessage(c.Channel, "There is no birthday on record for you")
		return
	}

	err = db.DeleteValue(append(bj.dbPath, "records"), c.User)
	if err != nil {
		bj.errorMsg(c, err, "cannot delete birthday")
		return
	}

	slack.SendMessage(c.Channel, "Birthday deleted!")

}
