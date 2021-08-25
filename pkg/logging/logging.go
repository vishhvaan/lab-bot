package logging

import (
	"io"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

const logFolder = "logs"
const logLevel = log.InfoLevel

func Setup() {
	log.SetLevel(logLevel)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Cannot stat current working directory")
	}

	fullLogFolder := path.Join(cwd, logFolder)
	if _, err := os.Stat(fullLogFolder); os.IsNotExist(err) {
		err = os.Mkdir(fullLogFolder, 0600)
		if err != nil {
			log.Fatal("Cannot make logging directory")
		}
	}

	logFile, err := os.OpenFile(logFolder, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal("Cannot open log file")
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}

func CreateNewLogger(prefix string, fileName string) {

}
