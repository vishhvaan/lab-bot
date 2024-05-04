package jobs

import (
	"strconv"
	"strings"
	"time"

	"github.com/vishhvaan/lab-bot/db"
	"github.com/vishhvaan/lab-bot/functions"
	"github.com/vishhvaan/lab-bot/scheduling"
	"github.com/vishhvaan/lab-bot/slack"
)

type birthdayJob struct {
	labJob
	dbPath           []string
	birthdaysdDbPath []string
	scheduling       scheduling.BirthdaySchedule
}

type birthday struct {
	date time.Time
	user string
}

func (bj *birthdayJob) init() {
	bj.labJob.init()

	bj.dbPath = append([]string{"jobs", "controller"}, bj.keyword)
	bj.birthdaysdDbPath = append(bj.dbPath, "birthdays")

	// ensure database is there or create database
	bj.scheduling.Sched = make(map[string]*scheduling.Schedule)
	bj.scheduling.DbPath = append(bj.dbPath, "scheduling")
	bj.checkCreateBucket()
	bj.checkBirthdaysExist()
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
	if !db.CheckBucketExists(bj.birthdaysdDbPath) {
		db.CreateBucket(bj.birthdaysdDbPath)
		bj.createBirthdayDBStructure()
	} else {
		// verify that all the months exist
	}

	if !db.CheckBucketExists(bj.scheduling.DbPath) {
		db.CreateBucket(bj.scheduling.DbPath)
	}

	return exists
}

func (bj *birthdayJob) createBirthdayDBStructure() error {
	for i := 1; i <= 12; i++ {
		err := db.CreateBucket(append(bj.birthdaysdDbPath, strconv.Itoa(i)))
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
