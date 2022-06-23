package functions

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	sysinfo "github.com/elastic/go-sysinfo"
)

func GetSysInfo() (si string) {
	var hostinfo strings.Builder
	h, err := sysinfo.Host()
	if err != nil {
		log.Error("Cannot parse host information.")
		return "Not Available"
	}

	hostinfo.WriteString("*" + h.Info().Hostname + "* running " + h.Info().OS.Family + ", " + h.Info().OS.Type + "/n" +
		"*Timezone*: " + h.Info().Timezone + "/n")

	hostinfo.WriteString("*Containerized*: ")
	if *h.Info().Containerized {
		hostinfo.WriteString("Yes")
	} else {
		hostinfo.WriteString("No")
	}
	hostinfo.WriteString("/n")

	c, err := h.CPUTime()
	if err != nil {
		log.Error("Cannot parse host CPU information.")
	} else {
		hostinfo.WriteString("*CPU Usage*: " + c.System.String() + "/n")
	}

	m, err := h.Memory()
	if err != nil {
		log.Error("Cannot parse host memory information.")
	} else {
		hostinfo.WriteString("*Memory Usage*: " + strconv.Itoa(int(m.Used)) + " used out of " + strconv.Itoa(int(m.Total)) + "/n")
	}

	return hostinfo.String()
}
