package jobs

import (
	"fmt"
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

// birthday status [@user]
func (bj *birthdayJob) birthdayStatus(c slack.CommandInfo) {
	isMention := func(tok string) bool {
		return strings.HasPrefix(tok, "<@") && strings.HasSuffix(tok, ">")
	}
	mentionToID := func(tok string) string {
		return strings.TrimSuffix(strings.TrimPrefix(tok, "<@"), ">")
	}

	// parse tokens after “birthday status”
	targetUser := c.User
	mentionSeen := false

	for _, tok := range c.Fields[2:] { // skip “birthday” “status”
		if isMention(tok) {
			if mentionSeen {
				slack.SendMessage(c.Channel,
					"please mention at most one user")
				return
			}
			targetUser = mentionToID(tok)
			mentionSeen = true
			continue
		}
		// any other token is unexpected
		slack.SendMessage(c.Channel,
			"usage: birthday status [@user]")
		return
	}

	// fetch birthday from DB
	b, err := db.ReadValue(append(bj.dbPath, "records"), targetUser)
	if err != nil {
		bj.errorMsg(c, err, "cannot read birthday from db")
		return
	}
	if b == nil {
		if targetUser == c.User {
			slack.SendMessage(c.Channel, "You have no birthday on record")
		} else {
			m := fmt.Sprintf("%s has no birthday on record", slack.GetUserName(targetUser))
			slack.SendMessage(c.Channel, m)

		}
		return
	}

	var bd time.Time
	if err := bd.UnmarshalJSON(b); err != nil {
		bj.errorMsg(c, err, "cannot decode birthday from db")
		return
	}

	// format & reply
	display := bd.Format("January 2") // show only month-day, ignore stored year

	if targetUser == c.User {
		slack.SendMessage(c.Channel,
			fmt.Sprintf("Your birthday on record is *%s*", display))
	} else {
		slack.SendMessage(c.Channel,
			fmt.Sprintf("%s's birthday on record is *%s*", slack.GetUserName(targetUser), display))
	}
}

// record a birthday:  “birthday record 10-24 [@user] [force]”
func (bj *birthdayJob) recordBirthday(c slack.CommandInfo) {
	loc := time.Now().Location()

	parseMonthDay := func(s string) (time.Time, bool) {
		var sep string
		if strings.Contains(s, "-") {
			sep = "-"
		} else if strings.Contains(s, "/") {
			sep = "/"
		} else {
			return time.Time{}, false
		}
		parts := strings.Split(s, sep)
		if len(parts) != 2 {
			return time.Time{}, false
		}
		month, errM := strconv.Atoi(parts[0])
		day, errD := strconv.Atoi(parts[1])
		if errM != nil || errD != nil ||
			month < 1 || month > 12 ||
			day < 1 || day > 31 {
			return time.Time{}, false
		}
		return time.Date(time.Now().Year(), time.Month(month), day, 0, 0, 0, 0, loc), true
	}

	isMention := func(tok string) bool {
		return strings.HasPrefix(tok, "<@") && strings.HasSuffix(tok, ">")
	}
	mentionToID := func(tok string) string {
		return strings.TrimSuffix(strings.TrimPrefix(tok, "<@"), ">")
	}

	// parse tokens after “birthday record”
	if len(c.Fields) < 3 {
		slack.SendMessage(c.Channel,
			"usage: birthday record <MM-DD | YYYY-MM-DD> [@user] [force]")
		return
	}

	targetUser := c.User // default to caller
	var dateToken string
	force := false
	mentionSeen := false

	for _, tok := range c.Fields[2:] {
		switch {
		case tok == "force":
			force = true

		case isMention(tok):
			if mentionSeen {
				slack.SendMessage(c.Channel,
					"Please mention at most one user. usage: birthday record <MM-DD | YYYY-MM-DD> [force]")
				return
			}
			targetUser = mentionToID(tok)
			mentionSeen = true

		default: // assume this is the date string
			if dateToken == "" {
				dateToken = tok
			} else {
				slack.SendMessage(c.Channel,
					"Too many date tokens; please supply only one. usage: birthday record <MM-DD | YYYY-MM-DD> [force]")
				return
			}
		}
	}

	if dateToken == "" {
		slack.SendMessage(c.Channel, "Birthday date is missing")
		return
	}

	// convert date string ➜ time.Time (midnight, current year)
	var newBD time.Time
	if bd, ok := parseMonthDay(dateToken); ok {
		newBD = bd
	} else {
		var err error
		newBD, err = dateparse.ParseAny(dateToken)
		if err != nil {
			go bj.logger.WithField("fields", c.Fields).
				WithError(err).Warn("cannot parse date")
			slack.PostMessage(c.Channel, "cannot parse date. usage: birthday record <MM-DD | YYYY-MM-DD> [force] -- "+err.Error())
			return
		}
	}
	newBD = newBD.Truncate(24 * time.Hour).In(loc)

	// db read / write
	b, err := db.ReadValue(append(bj.dbPath, "records"), targetUser)
	if err != nil {
		bj.errorMsg(c, err, "cannot read existing birthday from db")
		return
	}

	if b == nil || force {
		// save / overwrite
		byteBD, _ := newBD.MarshalJSON()
		if err = db.AddValue(append(bj.dbPath, "records"), targetUser, byteBD); err != nil {
			bj.errorMsg(c, err, "cannot record birthday to database")
			return
		}
		slack.React(c.TimeStamp, c.Channel, "tada")
		return
	}

	// duplicate handling
	var oldBD time.Time
	if err = oldBD.UnmarshalJSON(b); err != nil {
		bj.errorMsg(c, err, "cannot read existing birthday from db")
		return
	}

	if oldBD.Month() == newBD.Month() && oldBD.Day() == newBD.Day() {
		slack.SendMessage(c.Channel, "This birthday is already on record")
	} else {
		slack.SendMessage(c.Channel,
			"A different birthday is already on record; delete it first or use the 'force' flag")
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
