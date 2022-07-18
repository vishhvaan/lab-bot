package functions

import (
	"fmt"
	"strings"
	"time"

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

	hostinfo.WriteString("*" + h.Info().Hostname + "* running ")
	if h.Info().OS.Family == "" {
		if h.Info().OS.Type != "" {
			hostinfo.WriteString(h.Info().OS.Type + "\n")
		}
	} else {
		if h.Info().OS.Type == "" {
			hostinfo.WriteString(h.Info().OS.Family + "\n")
		} else {
			hostinfo.WriteString(h.Info().OS.Family + ", " + h.Info().OS.Type + "\n")
		}
	}
	hostinfo.WriteString("*Datetime*: " + time.Now().Format(time.UnixDate) + "\n")

	hostinfo.WriteString("*Containerized*: ")
	if *h.Info().Containerized {
		hostinfo.WriteString("Yes")
	} else {
		hostinfo.WriteString("No")
	}
	hostinfo.WriteString("\n")

	c, err := h.CPUTime()
	if err != nil {
		log.Error("Cannot parse host CPU information.")
	} else {
		hostinfo.WriteString("*Uptime*: " + c.System.String() + "\n")
	}

	m, err := h.Memory()
	if err != nil {
		log.Error("Cannot parse host memory information.")
	} else {
		hostinfo.WriteString("*Memory Usage*: " + ByteCountSI(m.Used) + " used out of " + ByteCountSI(m.Total) + "\n")
	}

	return hostinfo.String()
}

func ByteCountSI(b uint64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
