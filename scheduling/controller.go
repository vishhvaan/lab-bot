package scheduling

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	crondesc "github.com/lnquy/cron"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/db"
	"github.com/vishhvaan/lab-bot/slack"
)

type ControllerSchedule struct {
	Logger *log.Entry
	Sched  map[string]*Schedule
	DbPath []string
}

func (cs *ControllerSchedule) ContSet(id string, cronSched string, command slack.CommandInfo) (err error) {
	powerVal := command.Fields[1]
	if cs.Sched[powerVal] != nil && cs.Sched[powerVal].scheduler != nil && cs.Sched[powerVal].scheduler.IsRunning() {
		return errors.New("there exists a scheduled " + powerVal + " task")
	} else {
		_, err = cron.ParseStandard(cronSched)
		if err != nil {
			return err
		}

		s := gocron.NewScheduler(time.Now().Local().Location())

		name := command.Fields[0] + " " + command.Fields[2]
		s.Cron(cronSched).Tag(powerVal).Do(func(command slack.CommandInfo, id string, name string, channel string) {
			t := "[" + id + "] Executing " + name
			slack.MessageChan <- slack.MessageInfo{
				ChannelID: channel,
				Text:      t,
			}
			slack.CommandChan <- command
		}, command, id, name, command.Channel)
		s.StartAsync()

		record := scheduleRecord{
			id:      id,
			name:    name,
			cronExp: cronSched,
			command: command,
		}
		cs.writeSchedtoDB(record)

		sch := &Schedule{
			scheduleRecord: record,
			scheduler:      s,
			logger:         cs.Logger.WithField("job", name),
		}

		if err == nil {
			cs.Sched[powerVal] = sch
		}
		return err
	}
}

func (cs *ControllerSchedule) ContRemove(command slack.CommandInfo) (err error) {
	powerVal := command.Fields[1]
	if cs.Sched[powerVal] != nil && cs.Sched[powerVal].scheduler != nil && cs.Sched[powerVal].scheduler.IsRunning() {
		cs.Sched[powerVal].scheduler.Stop()
		// schedChan <- cs.onSched
		cs.deleteSchedfromDB(cs.Sched[powerVal].scheduleRecord)
		delete(cs.Sched, powerVal)
		return nil
	} else {
		return errors.New("there is no scheduled " + powerVal + " task")
	}
}

func (cs *ControllerSchedule) ContGetSchedulingStatus() string {
	var status strings.Builder
	exprDesc, err := crondesc.NewDescriptor()
	if err != nil {
		message := "descriptor failed to start up"
		cs.Logger.WithField("err", err).Error(message)
		return "*Scheduling*: " + message
	}

	for key, schedule := range cs.Sched {
		if schedule != nil && schedule.scheduler != nil && schedule.scheduler.IsRunning() {
			status.WriteString("*Scheduled " + strings.Title(key) + "*: ")
			onText, err := exprDesc.ToDescription(schedule.cronExp, crondesc.Locale_en)
			if err != nil {
				message := "could not generate plain text for scheduled " + key
				cs.Logger.WithField("err", err).Error(message)
				status.WriteString(message)
			} else {
				status.WriteString(onText)
			}
			status.WriteString("\n")
		}
	}

	if status.Len() == 0 {
		status.WriteString("*Scheduling*: Not setup")
	}

	return status.String()
}

func (cj *ControllerSchedule) LoadDBSched() (records []scheduleRecord, err error) {

	return err
}

func (cs *ControllerSchedule) writeSchedtoDB(record scheduleRecord) (err error) {
	buf, err := json.Marshal(record)
	if err != nil {
		cs.Logger.WithFields(log.Fields{
			"err":    err,
			"record": record,
		}).Error("cannot create convert struct to json")
		return err
	}

	err = db.AddValue(cs.DbPath, record.id, buf)
	return err
}

func (cs *ControllerSchedule) deleteSchedfromDB(record scheduleRecord) (err error) {

	return err
}
