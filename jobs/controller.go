package jobs

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/db"
	"github.com/vishhvaan/lab-bot/functions"
	"github.com/vishhvaan/lab-bot/scheduling"
	"github.com/vishhvaan/lab-bot/slack"
)

const controllerIDLen = 6

type controllerJob struct {
	labJob
	machineName string
	powerState  bool
	lastPowerOn time.Time
	device      any
	customInit  func() (err error)
	customOn    func() (err error)
	customOff   func() (err error)
	logger      *log.Entry
	scheduling  scheduling.ControllerSchedule
	dbPath      []string
	controller
}

type controller interface {
	init()
	TurnOn(c slack.CommandInfo)
	TurnOff(c slack.CommandInfo)
	getPowerStatus(c slack.CommandInfo)
	commandProcessor(c slack.CommandInfo)
}

func (cj *controllerJob) init() {
	cj.labJob.init()

	cj.dbPath = append([]string{"jobs", "controller"}, cj.keyword)

	var message string
	err := cj.customInit()
	if err != nil {
		message = "Couldn't load " + cj.name
		cj.logger.WithField("err", err).Error(message)
	} else {
		message = cj.name + " loaded"
		cj.logger.Info(message)
	}
	slack.MessageChan <- slack.MessageInfo{
		Text: message,
	}

	cj.scheduling.Sched = make(map[string]*scheduling.Schedule)
	cj.scheduling.DbPath = append(cj.dbPath, "scheduling")
	if cj.checkCreateBucket() {
		cj.loadSchedsFromDB()
		cj.loadPowerStateFromDB()
	} else {
		cj.updatePowerStateInDB()
	}

}

func (cj *controllerJob) checkCreateBucket() (exists bool) {
	exists = db.CheckBucketExists(cj.scheduling.DbPath)
	if !exists {
		db.CreateBucket(cj.scheduling.DbPath)
	}
	return exists
}

func (cj *controllerJob) loadSchedsFromDB() (err error) {
	records, err := cj.scheduling.LoadSchedsfromDB()
	if err != nil {
		message := "Cannot load schedules from database"
		slack.MessageChan <- slack.MessageInfo{
			Text: message,
		}
		cj.logger.WithFields(log.Fields{
			"err":  err,
			"path": cj.scheduling.DbPath,
		}).Error(message)
		return err
	}

	for _, record := range records {
		powerVal := record.Command.Fields[2]
		e := cj.scheduling.ContSet(record.ID, record.CronExp, record.Command, false)
		if e != nil {
			cj.errorMsg(record.Command.Fields, record.Command.Channel, err.Error())
			err = e
		} else {
			slack.MessageChan <- slack.MessageInfo{
				Text: "_Loaded scheduled power " + powerVal + " task from the database._",
			}
			cj.logger.WithFields(log.Fields{
				"id":       record.ID,
				"powerval": powerVal,
				"path":     cj.scheduling.DbPath,
			}).Info("Loaded scheduled task from the database")
		}
	}
	return err
}

func (cj *controllerJob) updatePowerStateInDB() (err error) {
	powerString := "off"
	if cj.powerState {
		powerString = "on"
	}

	err = db.AddValue(cj.dbPath, "powerState", []byte(powerString))

	return err
}

func (cj *controllerJob) loadPowerStateFromDB() (err error) {
	v, err := db.ReadValue(cj.dbPath, "powerState")
	powerString := string(v[:])

	if powerString == "on" {
		cj.powerState = true
	} else {
		cj.powerState = false
	}

	return err
}

func (cj *controllerJob) TurnOn(c slack.CommandInfo) {
	if commandCheck(c, 2, slack.MessageChan, cj.logger) {
		if cj.powerState {
			message := "The " + cj.machineName + " is already on"
			go cj.logger.Info(message)
			slack.MessageChan <- slack.MessageInfo{
				Text:      message,
				ChannelID: c.Channel,
			}
		} else {
			err := cj.customOn()
			cj.lastPowerOn = time.Now()
			cj.slackPowerResponse(true, err, c)
			cj.updatePowerStateInDB()
		}
	}
}

func (cj *controllerJob) TurnOff(c slack.CommandInfo) {
	if commandCheck(c, 2, slack.MessageChan, cj.logger) {
		if !cj.powerState {
			message := "The " + cj.machineName + " is already off"
			go cj.logger.Info(message)
			slack.MessageChan <- slack.MessageInfo{
				Text:      message,
				ChannelID: c.Channel,
			}
		} else {
			err := cj.customOff()
			cj.slackPowerResponse(false, err, c)
			cj.updatePowerStateInDB()
		}
	}
}

func (cj *controllerJob) slackPowerResponse(status bool, err error, c slack.CommandInfo) {
	statusString := "off"
	if status {
		statusString = "on"
	}
	if err != nil {
		message := "Couldn't turn " + statusString + " the " + cj.machineName
		go cj.logger.WithField("err", err).Error(message)
		slack.MessageChan <- slack.MessageInfo{
			Text: message,
		}
	} else {
		cj.powerState = status
		message := "Turned " + statusString + " the " + cj.machineName
		go cj.logger.Info(message)
		if c.TimeStamp != "" {
			slack.MessageChan <- slack.MessageInfo{
				Type:      "react",
				Timestamp: c.TimeStamp,
				ChannelID: c.Channel,
				Text:      "ok_hand",
			}
		}
		slack.MessageChan <- slack.MessageInfo{
			Text: message,
		}
	}
}

func (cj *controllerJob) getPowerStatus(c slack.CommandInfo) {
	if commandCheck(c, 2, slack.MessageChan, cj.logger) {
		message := "The " + cj.machineName + " is "
		if cj.powerState {
			uptime := time.Since(cj.lastPowerOn).Round(time.Second)
			message += "*on*\nUptime: " + fmt.Sprint(uptime)
		} else {
			message += "*off*"
		}
		message += "\n" + cj.scheduling.ContGetSchedulingStatus()

		slack.MessageChan <- slack.MessageInfo{
			ChannelID: c.Channel,
			Text:      message,
		}
	}
}

func (cj *controllerJob) errorMsg(fields []string, channel string, message string) {
	go cj.logger.WithField("fields", fields).Warn(message)
	slack.MessageChan <- slack.MessageInfo{
		ChannelID: channel,
		Text:      message,
	}
}

func (cj *controllerJob) sendMsg(channel string, message string) {
	go cj.logger.Info(message)
	slack.MessageChan <- slack.MessageInfo{
		ChannelID: channel,
		Text:      message,
	}
}

func (cj *controllerJob) commandProcessor(c slack.CommandInfo) {
	if cj.active {
		controllerActions := map[string]action{
			"on":       cj.TurnOn,
			"off":      cj.TurnOff,
			"status":   cj.getPowerStatus,
			"schedule": cj.scheduleHandler,
		}
		if len(c.Fields) == 1 {
			cj.getPowerStatus(c)
		} else {
			k := functions.GetKeys(controllerActions)
			subcommand := strings.ToLower(c.Fields[1])
			if functions.Contains(k, subcommand) {
				f := controllerActions[subcommand]
				f(c)
			} else {
				cj.errorMsg(c.Fields, c.Channel, "I'm not sure what you sayin")
			}
		}
	} else {
		slack.MessageChan <- slack.MessageInfo{
			ChannelID: c.Channel,
			Text:      "The " + cj.name + " is disabled",
		}
	}
}

func (cj *controllerJob) scheduleHandler(c slack.CommandInfo) {
	schedulingActions := map[string]action{
		"on":     cj.sched,
		"off":    cj.sched,
		"status": cj.sendSchedulingStatus,
	}
	if len(c.Fields) == 2 {
		cj.sendSchedulingStatus(c)
	} else if len(c.Fields) > 2 {
		k := functions.GetKeys(schedulingActions)
		subcommand := strings.ToLower(c.Fields[2])
		if functions.Contains(k, subcommand) {
			f := schedulingActions[subcommand]
			f(c)
		} else {
			cj.errorMsg(c.Fields, c.Channel, "I'm not sure what you sayin")
		}
	}
}

func (cj *controllerJob) sched(c slack.CommandInfo) {
	powerVal := c.Fields[2]
	// keyword = c.Fields[0]
	if len(c.Fields) >= 4 {
		if c.Fields[3] == "set" && len(c.Fields) > 4 {
			cronExp := strings.Join(c.Fields[4:], " ")
			idNum, err := db.IncrementBucketInteger(cj.scheduling.DbPath)
			idString := strconv.Itoa(idNum) + c.Fields[0] + "controller"
			id := functions.SHA256Sum(idString, controllerIDLen)
			if err != nil {
				cj.errorMsg(c.Fields, c.Channel, "couldn't get ID for schedule")
			}
			err = cj.scheduling.ContSet(id, cronExp, c, true)
			if err != nil {
				cj.errorMsg(c.Fields, c.Channel, err.Error())
			} else {
				cj.sendMsg(c.Channel, "_Successfully scheduled power "+powerVal+" task._\n"+cj.scheduling.ContGetSchedulingStatus())
			}
			return
		} else if c.Fields[3] == "remove" && len(c.Fields) == 4 {
			err := cj.scheduling.ContRemove(c)
			if err != nil {
				cj.errorMsg(c.Fields, c.Channel, err.Error())
			} else {
				cj.sendMsg(c.Channel, "_Successfully removed power "+powerVal+" task._\n"+cj.scheduling.ContGetSchedulingStatus())
			}
			return
		}
	}
	cj.errorMsg(c.Fields, c.Channel, "Malformed scheduling command")
}

func (cj *controllerJob) sendSchedulingStatus(c slack.CommandInfo) {
	slack.MessageChan <- slack.MessageInfo{
		ChannelID: c.Channel,
		Text:      cj.scheduling.ContGetSchedulingStatus(),
	}
}
