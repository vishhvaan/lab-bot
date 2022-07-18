package scheduling

import (
	"errors"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	crondesc "github.com/lnquy/cron"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/pkg/slack"
)

type ControllerSchedule struct {
	Logger   *log.Entry
	onSched  *Schedule
	offSched *Schedule
}

func (cs *ControllerSchedule) ContSetOn(cronSched string, channel string, keyword string, m chan slack.MessageInfo, c chan slack.CommandInfo) (err error) {
	if cs.onSched.scheduler != nil && cs.onSched.scheduler.IsRunning() {
		return errors.New("there exists a scheduled on task")
	} else {
		s, err := cs.contSet(cronSched, channel, keyword, m, c, "on")
		if err == nil {
			cs.onSched = s
		}
		return err
	}
}

func (cs *ControllerSchedule) ContSetOff(cronSched string, channel string, keyword string, m chan slack.MessageInfo, c chan slack.CommandInfo) (err error) {
	if cs.offSched.scheduler != nil && cs.offSched.scheduler.IsRunning() {
		return errors.New("there exists a scheduled off task")
	} else {
		s, err := cs.contSet(cronSched, channel, keyword, m, c, "off")
		if err == nil {
			cs.offSched = s
		}
		return err
	}
}

func (cs *ControllerSchedule) contSet(cronSched string, channel string, keyword string, m chan slack.MessageInfo, c chan slack.CommandInfo, powerVal string) (sched *Schedule, err error) {
	_, err = cron.ParseStandard(cronSched)
	if err != nil {
		return &Schedule{}, err
	}

	s := gocron.NewScheduler(time.Now().Local().Location())
	command := slack.CommandInfo{
		Fields:  []string{keyword, powerVal},
		Channel: channel,
	}

	name := keyword + " " + powerVal
	id := generateID()
	s.Cron(cronSched).Tag(powerVal).Do(func(m chan slack.MessageInfo, c chan slack.CommandInfo, command slack.CommandInfo, id string, name string, channel string) {
		t := "[" + id + "] Executing " + name
		m <- slack.MessageInfo{
			ChannelID: channel,
			Text:      t,
		}
		c <- command
	}, m, c, command, id, name, channel)
	s.StartAsync()

	sch := &Schedule{
		id:        id,
		name:      name,
		cronExp:   cronSched,
		channel:   channel,
		scheduler: s,
		logger:    cs.Logger.WithField("job", name),
	}
	schedChan <- sch
	return sch, nil
}

func (cs *ControllerSchedule) ContRemoveOn() (err error) {
	if cs.onSched.scheduler.IsRunning() {
		cs.onSched.scheduler.Stop()
		schedChan <- cs.onSched
		cs.onSched = nil
		return nil
	} else {
		return errors.New("there is no scheduled on task")
	}
}

func (cs *ControllerSchedule) ContRemoveOff() (err error) {
	if cs.offSched.scheduler.IsRunning() {
		cs.offSched.scheduler.Stop()
		schedChan <- cs.offSched
		cs.offSched = nil
		return nil
	} else {
		return errors.New("there is no scheduled off task")
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

	if cs.onSched != nil && cs.onSched.scheduler != nil && cs.onSched.scheduler.IsRunning() {
		status.WriteString("*Scheduled On*: ")
		onText, err := exprDesc.ToDescription(cs.onSched.cronExp, crondesc.Locale_en)
		if err != nil {
			message := "could not generate plain text for scheduled on"
			cs.Logger.WithField("err", err).Error(message)
			status.WriteString(message)
		} else {
			status.WriteString(onText)
		}
		status.WriteString("\n")
	}

	if cs.offSched != nil && cs.offSched.scheduler != nil && cs.offSched.scheduler.IsRunning() {
		status.WriteString("*Scheduled Off*: ")
		onText, err := exprDesc.ToDescription(cs.offSched.cronExp, crondesc.Locale_en)
		if err != nil {
			message := "could not generate plain text for scheduled off"
			cs.Logger.WithField("err", err).Error(message)
			status.WriteString(message)
		} else {
			status.WriteString(onText)
		}
		status.WriteString("\n")
	}

	if status.Len() == 0 {
		status.WriteString("*Scheduling*: Not setup")
	}

	return status.String()
}
