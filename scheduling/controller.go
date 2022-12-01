package scheduling

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	crondesc "github.com/lnquy/cron"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/db"
	"github.com/vishhvaan/lab-bot/functions"
	"github.com/vishhvaan/lab-bot/slack"
)

type ControllerSchedule struct {
	Set                   bool
	powerMessageChannel   string
	powerMessageTimestamp string
	Logger                *log.Entry
	Sched                 map[string]*Schedule
	DbPath                []string
}

func (cs *ControllerSchedule) ContSet(id string, cronSched string, command slack.CommandInfo, newSched bool) (err error) {
	powerVal := command.Fields[2]
	if cs.Sched[powerVal] != nil && cs.Sched[powerVal].scheduler != nil && cs.Sched[powerVal].scheduler.IsRunning() {
		return errors.New("there exists a scheduled " + powerVal + " task")
	} else {
		_, err = cron.ParseStandard(cronSched)
		if err != nil {
			return err
		}

		s := gocron.NewScheduler(time.Now().Local().Location())

		name := command.Fields[0] + " " + command.Fields[2]
		s.Cron(cronSched).Tag(powerVal).Do(func(command slack.CommandInfo, id string, name string) {
			slack.CommandChan <- slack.CommandInfo{
				Fields:  []string{command.Fields[0], command.Fields[2]},
				Channel: command.Channel,
			}
		}, command, id, name)
		s.StartAsync()

		record := scheduleRecord{
			ID:      id,
			Name:    name,
			CronExp: cronSched,
			Command: command,
		}

		if newSched {
			err = cs.writeSchedtoDB(record)
			l := cs.Logger.WithFields(log.Fields{
				"id":   id,
				"name": name,
			})
			if err != nil {
				l.WithError(err).Error("Cannot add schedule to db")
			} else {
				l.Info("Added schedule to db")
			}
		}

		sch := &Schedule{
			scheduleRecord: record,
			scheduler:      s,
			logger:         cs.Logger.WithField("job", name),
		}

		if err == nil {
			cs.Sched[powerVal] = sch
			cs.Set = true
		}
		return err
	}
}

func (cs *ControllerSchedule) ContRemove(command slack.CommandInfo) (err error) {
	powerVal := command.Fields[2]
	if cs.Sched[powerVal] != nil && cs.Sched[powerVal].scheduler != nil && cs.Sched[powerVal].scheduler.IsRunning() {
		cs.Sched[powerVal].scheduler.Stop()
		// schedChan <- cs.onSched

		err = cs.deleteSchedfromDB(cs.Sched[powerVal].scheduleRecord)
		if err != nil {
			cs.Logger.WithField("id", cs.Sched[powerVal].scheduleRecord.ID).Error("Could not delete schedule from database")
		} else {
			cs.Logger.WithField("id", cs.Sched[powerVal].scheduleRecord.ID).Info("Deleted schedule from database")
		}
		delete(cs.Sched, powerVal)
		if len(cs.Sched) == 0 {
			cs.Set = false
			cs.DeletePowerMessage()
		}
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
			onText, err := exprDesc.ToDescription(schedule.CronExp, crondesc.Locale_en)
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

func (cs *ControllerSchedule) LoadSchedsfromDB() (records []scheduleRecord, err error) {
	_, values, err := db.GetAllKeysValues(append(cs.DbPath, "records"))
	if err != nil {
		return records, err
	}

	for _, value := range values {
		var record scheduleRecord
		err = json.Unmarshal(value, &record)
		cs.Logger.WithField("id", record.ID).Info("Parsed schedule in database")
		records = append(records, record)
	}
	return records, err
}

func (cs *ControllerSchedule) writeSchedtoDB(record scheduleRecord) (err error) {
	recordsPath := append(cs.DbPath, "records")
	value, err := db.ReadValue(recordsPath, record.ID)
	if value != nil {
		cs.Logger.WithFields(log.Fields{
			"path": cs.DbPath,
			"id":   record.ID,
		}).Error("Schedule already exists with this id")
		return err
	}

	buf, err := json.Marshal(record)
	if err != nil {
		cs.Logger.WithFields(log.Fields{
			"err":    err,
			"record": record,
		}).Error("Cannot create convert struct to json")
		return err
	}

	err = db.AddValue(recordsPath, record.ID, buf)
	return err
}

func (cs *ControllerSchedule) deleteSchedfromDB(record scheduleRecord) (err error) {
	err = db.DeleteValue(append(cs.DbPath, "records"), record.ID)
	return err
}

func (cs *ControllerSchedule) LoadPowerMessagefromDB() error {
	var readMessageChannel []byte
	readTimestamp, err := db.ReadValue(cs.DbPath, "PowerMessageTimestamp")
	if err == nil {
		if readTimestamp != nil {
			readMessageChannel, err = db.ReadValue(cs.DbPath, "PowerMessageChannel")
			if err == nil {
				if readMessageChannel == nil {
					cs.Logger.WithFields(log.Fields{
						"err": err,
					}).Warn("Cannot load power message channel")
					return errors.New("channel doesn't exist")
				}
			} else {
				return err
			}
		} else {
			cs.Logger.WithFields(log.Fields{
				"err": err,
			}).Warn("Cannot load power message timestamp")
			return errors.New("timestamp doesn't exist")
		}
	}

	cs.powerMessageTimestamp = string(readTimestamp[:])
	cs.powerMessageChannel = string(readMessageChannel[:])

	return err
}

func (cs *ControllerSchedule) PostPowerMessage(channel string, name string, status string) (err error) {
	cs.powerMessageChannel = channel
	cs.powerMessageTimestamp, err = slack.PostMessage(channel, name+": "+status)
	if err == nil {
		slack.PinMessage(cs.powerMessageChannel, cs.powerMessageTimestamp)
		db.AddValue(cs.DbPath, "PowerMessageTimestamp", []byte(cs.powerMessageTimestamp))
		db.AddValue(cs.DbPath, "PowerMessageChannel", []byte(cs.powerMessageChannel))

		cs.DeleteOtherPowerMessage(name, cs.powerMessageTimestamp)

		cs.Logger.WithFields(log.Fields{
			"channel":   cs.powerMessageChannel,
			"timestamp": cs.powerMessageTimestamp,
		}).Info("Posted new power message")
	}
	return err
}

func (cs *ControllerSchedule) DeletePowerMessage() error {
	err := slack.DeleteMessage(cs.powerMessageChannel, cs.powerMessageTimestamp)
	if err == nil {
		db.DeleteValue(cs.DbPath, "PowerMessageTimestamp")
		db.DeleteValue(cs.DbPath, "PowerMessageChannel")
		cs.powerMessageChannel = ""
		cs.powerMessageTimestamp = ""
	}
	return err
}

func (cs *ControllerSchedule) DeleteOtherPowerMessage(name string, timestamp string) error {
	if cs.powerMessageChannel == "" || cs.powerMessageTimestamp == "" {
		errorMsg := "Power message channel is undefined, cannot delete power messages."
		cs.Logger.Error(errorMsg)
		return errors.New(errorMsg)
	}

	pinnedMessages, err := slack.ListPins(cs.powerMessageChannel)
	if err == nil {
		timestamps := functions.GetKeys(pinnedMessages)
		var numDeleted int
		if len(timestamps) != 0 {
			for _, timestamp := range timestamps {
				if timestamp != cs.powerMessageTimestamp && strings.Contains(pinnedMessages[timestamp], name) {
					slack.DeleteMessage(cs.powerMessageChannel, timestamp)
					numDeleted++
				}
			}
			cs.Logger.WithFields(log.Fields{
				"channel": cs.powerMessageChannel,
			}).Info(fmt.Sprintf("Deleted %d power messages.", numDeleted))
		}
	} else {
		cs.Logger.WithError(err).Warn("Couldn't delete power messages")
	}
	return err
}

func (cs *ControllerSchedule) ModifyPowerMessage(name string, status string) error {
	err := slack.ModifyMessage(cs.powerMessageChannel, cs.powerMessageTimestamp, name+": "+status)
	if err != nil {
		cs.Logger.WithFields(log.Fields{
			"channel":   cs.powerMessageChannel,
			"timestamp": cs.powerMessageTimestamp,
		}).Warn("Couldn't modify power message")
	} else {
		cs.Logger.WithFields(log.Fields{
			"channel":   cs.powerMessageChannel,
			"timestamp": cs.powerMessageTimestamp,
		}).Info("Modified power message")

		cs.DeleteOtherPowerMessage(name, cs.powerMessageTimestamp)
	}
	return err
}
