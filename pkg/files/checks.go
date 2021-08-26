package files

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func CheckFile(file string) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"file": file,
			"error": err,
		}).Fatal("File doesn't exist.")

	}
}
