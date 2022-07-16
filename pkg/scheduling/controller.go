package scheduling

import log "github.com/sirupsen/logrus"

type ControllerSchedule struct {
	enabled  bool
	channel  string
	logger   *log.Entry
	onSched  *Schedule
	offSched *Schedule
}

func (cs *ControllerSchedule) SetOn(cronSched string) {

}

func (cs *ControllerSchedule) SetOff(cronSched string) {

}

func (cs *ControllerSchedule) RemoveOn() {

}

func (cs *ControllerSchedule) RemoveOff() {

}

func (cs *ControllerSchedule) GetSchedulingStatus() string {

}
