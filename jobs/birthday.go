package jobs

import (
	"encoding/json"
	"strconv"
	"strings"

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

// db organization birthdays/numeric month/key = day, value = user(s) slice

func (bj *birthdayJob) checkCreateBucket() (exists bool) {
	if !db.CheckBucketExists(bj.birthdayDbPath) {
		db.CreateBucket(bj.birthdayDbPath)
	}
	bj.createBirthdayDBStructure()

	if !db.CheckBucketExists(bj.scheduling.DbPath) {
		db.CreateBucket(bj.scheduling.DbPath)
	}

	return exists
}

func (bj *birthdayJob) createBirthdayDBStructure() (err error) {
	for i := 1; i <= 12; i++ {
		monthPath := append(bj.birthdayDbPath, strconv.Itoa(i))
		if !db.CheckBucketExists(monthPath) {
			err = db.CreateBucket(monthPath)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (bj *birthdayJob) errorMsg(fields []string, channel string, message string) {
	go bj.logger.WithField("fields", fields).Warn(message)
	slack.PostMessage(channel, message)
}

func (bj *birthdayJob) numerateBirthdays() (numBirthdays int, err error) {
	for i := 1; i <= 12; i++ {
		monthPath := append(bj.birthdayDbPath, strconv.Itoa(i))
		_, values, err := db.GetAllKeysValues(monthPath)
		if err != nil {
			return -1, err
		}

		for _, dayBirthdays := range values {
			var users []string
			err = json.Unmarshal(dayBirthdays, &users)
			if err != nil {
				return -1, err
			}
			numBirthdays += len(users)
		}
	}

	return numBirthdays, nil
}
