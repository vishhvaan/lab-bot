package jobs

import (
	"encoding/json"
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
	powerState  string
	lastPowerOn time.Time
	device      any
	customInit  func() (err error)
	customOn    func() (err error)
	customOff   func() (err error)
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
	slack.Message(message)

	cj.scheduling.Sched = make(map[string]*scheduling.Schedule)
	cj.scheduling.DbPath = append(cj.dbPath, "scheduling")
	if cj.checkCreateBucket() {
		cj.loadSchedsFromDB()
		cj.loadPowerStateFromDB()
	} else {
		cj.updatePowerStateInDB()
	}
}

func (cj *controllerJob) commandProcessor(c slack.CommandInfo) {
	if cj.active {
		controllerActions := map[string]action{
			"on":       cj.TurnOn,
			"off":      cj.TurnOff,
			"status":   cj.getPowerStatus,
			"schedule": cj.scheduleHandler,
			"force":    cj.forcePower,
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
		slack.PostMessage(c.Channel, "The "+cj.name+" is disabled")
	}
}

func (cj *controllerJob) checkCreateBucket() (exists bool) {
	exists = db.CheckBucketExists(cj.scheduling.DbPath)
	if !exists {
		db.CreateBucket(cj.scheduling.DbPath)
		db.CreateBucket(append(cj.scheduling.DbPath, "records"))
	}
	return exists
}

func (cj *controllerJob) loadSchedsFromDB() (err error) {
	records, err := cj.scheduling.LoadSchedsfromDB()
	if err != nil {
		message := "Cannot load schedules from database"
		slack.Message(message)
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
			slack.Message("_Loaded scheduled power " + powerVal + " task from the database._")
			cj.logger.WithFields(log.Fields{
				"id":       record.ID,
				"powerval": powerVal,
				"path":     cj.scheduling.DbPath,
			}).Info("Loaded scheduled task from the database")
		}
	}

	if len(records) != 0 {
		err = cj.scheduling.LoadPowerMessagefromDB()
		if err != nil {
			cj.logger.Warn("Couldn't load power message from db, posting new one")
			cj.scheduling.PostPowerMessage(records[0].Command.Channel, cj.name, cj.powerState)
		} else {
			cj.logger.Info("Found power message in db, testing")
			err := cj.scheduling.ModifyPowerMessage(cj.name, "testing")
			if err != nil {
				cj.logger.Warn("Power message testing failed, deleting and posting new one")
				cj.scheduling.DeletePowerMessage()
				cj.scheduling.PostPowerMessage(records[0].Command.Channel, cj.name, cj.powerState)
			} else {
				cj.logger.Info("Power message testing succeeded")
			}
		}
	}

	return err
}

func (cj *controllerJob) updatePowerStateInDB() (err error) {
	buf, err := json.Marshal(cj.lastPowerOn)
	if err != nil {
		cj.logger.WithFields(log.Fields{
			"err":         err,
			"lastPowerOn": cj.lastPowerOn.String(),
		}).Error("Cannot create convert power on time to json")
		return err
	}

	err = db.AddValue(cj.dbPath, "powerState", []byte(cj.powerState))
	if err == nil {
		err = db.AddValue(cj.dbPath, "lastPowerOn", buf)
	}

	return err
}

func (cj *controllerJob) loadPowerStateFromDB() (err error) {
	v, err := db.ReadValue(cj.dbPath, "powerState")
	cj.powerState = string(v[:])

	if err == nil {
		var buf []byte
		var lastPowerOn time.Time
		buf, err = db.ReadValue(cj.dbPath, "lastPowerOn")
		if err == nil {
			err = json.Unmarshal(buf, &lastPowerOn)
			if err != nil {
				cj.logger.WithField("machine", cj.machineName).Error("Cannot unmarshal lastPowerOn from db")
			} else {
				cj.lastPowerOn = lastPowerOn
			}

			if cj.scheduling.Set {
				cj.scheduling.ModifyPowerMessage(cj.name, cj.powerState)
			}
		}

	}

	return err
}

func (cj *controllerJob) powerControl(c slack.CommandInfo, powerState string, force bool) {
	powerFunctions := map[string]func() error{
		"on":  cj.customOn,
		"off": cj.customOff,
	}
	numParams := 2
	if force {
		numParams = 3
	}
	if commandCheck(c, numParams, cj.logger) && !force {
		if cj.powerState == powerState {
			message := "The " + cj.machineName + " is already " + powerState
			go cj.logger.Info(message)
			slack.PostMessage(c.Channel, message)
		} else {
			err := powerFunctions[powerState]()
			cj.lastPowerOn = time.Now()
			cj.powerState = powerState
			if cj.scheduling.Set {
				cj.scheduling.ModifyPowerMessage(cj.name, cj.powerState)
			}
			cj.slackPowerResponse(cj.powerState, err, c)
			cj.updatePowerStateInDB()
		}
	}
}

func (cj *controllerJob) TurnOn(c slack.CommandInfo) {
	cj.powerControl(c, "on", false)
}

func (cj *controllerJob) TurnOff(c slack.CommandInfo) {
	cj.powerControl(c, "on", false)
}

func (cj *controllerJob) turnOnForce(c slack.CommandInfo) {
	cj.powerControl(c, "on", true)
}

func (cj *controllerJob) turnOffForce(c slack.CommandInfo) {
	cj.powerControl(c, "on", true)
}

func (cj *controllerJob) forcePower(c slack.CommandInfo) {
	forceActions := map[string]action{
		"on":  cj.turnOnForce,
		"off": cj.turnOffForce,
	}
	if len(c.Fields) == 2 {
		cj.errorMsg(c.Fields, c.Channel, "Force on or off?")
	} else if len(c.Fields) > 2 {
		k := functions.GetKeys(forceActions)
		subcommand := strings.ToLower(c.Fields[2])
		if functions.Contains(k, subcommand) {
			f := forceActions[subcommand]
			f(c)
		} else {
			cj.errorMsg(c.Fields, c.Channel, "I'm not sure what you sayin")
		}
	}
}

func (cj *controllerJob) slackPowerResponse(status string, err error, c slack.CommandInfo) {
	if err != nil {
		message := "Couldn't turn " + status + " the " + cj.machineName
		go cj.logger.WithField("err", err).Error(message)
		slack.Message(message)
	} else {
		message := "Turned " + status + " the " + cj.machineName
		go cj.logger.Info(message)
		if c.TimeStamp != "" {
			slack.React(c.TimeStamp, c.Channel, "ok_hand")
		}
		slack.Message(message)
	}
}

func (cj *controllerJob) getPowerStatus(c slack.CommandInfo) {
	if commandCheck(c, 2, cj.logger) {
		message := "The " + cj.machineName + " is "
		if cj.powerState == "on" {
			uptime := time.Since(cj.lastPowerOn).Round(time.Second)
			message += "*on*\nUptime: " + fmt.Sprint(uptime)
		} else {
			message += "*off*"
		}
		message += "\n" + cj.scheduling.ContGetSchedulingStatus()

		slack.PostMessage(c.Channel, message)
	}
}

func (cj *controllerJob) errorMsg(fields []string, channel string, message string) {
	go cj.logger.WithField("fields", fields).Warn(message)
	slack.PostMessage(channel, message)
}

func (cj *controllerJob) sendMsg(channel string, message string) {
	go cj.logger.Info(message)
	slack.PostMessage(channel, message)
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

			newSched := cj.scheduling.Set

			err = cj.scheduling.ContSet(id, cronExp, c, true)
			if err != nil {
				cj.errorMsg(c.Fields, c.Channel, err.Error())
			} else {
				cj.sendMsg(c.Channel, "_Successfully scheduled power "+powerVal+" task._\n"+cj.scheduling.ContGetSchedulingStatus())
				if !newSched {
					cj.scheduling.PostPowerMessage(c.Channel, cj.name, cj.powerState)
				}
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
	slack.PostMessage(c.Channel, cj.scheduling.ContGetSchedulingStatus())
}
