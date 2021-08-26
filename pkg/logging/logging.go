package logging

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/TheoryDivision/lab-bot/pkg/files"
)

const logFolder = "logs"
const logLevel = log.InfoLevel

func Setup() {
	log.SetLevel(logLevel)

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "Jan 02 15:04:05",
		FullTimestamp:   true,
	})

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Cannot stat current working directory")
	}

	logFolderFull, err := files.CreateFolder(cwd, logFolder)
	if err != nil {
		log.Fatal("Cannot open log folder")
	}

	logFile, err := files.OpenFile(logFolderFull, "main.log")
	if err != nil {
		log.Fatal("Cannot open log file")
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}

func CreateNewLogger(prefix string, fileName string) *log.Entry {
	var logger = log.New()

	logger.SetFormatter(&log.TextFormatter{
		TimestampFormat: "Jan 02 15:04:05",
		FullTimestamp:   true,
	})

	return logger.WithField("app", prefix)
}
