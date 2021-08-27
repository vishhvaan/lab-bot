package logging

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/vishhvaan/lab-bot/pkg/files"
)

const logFolder = "logs"
const logExt = ".log"
const logLevel = log.InfoLevel

func Setup() {
	log.SetLevel(logLevel)

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "Jan 02 15:04:05.000",
		FullTimestamp:   true,
		ForceColors:     true,
	})

	logPath := CreateLogFolder()
	logFile := CreateLogFile(logPath, "main")
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}

func CreateNewLogger(prefix string, filename string) *log.Entry {
	var logger = log.New()
	logger.SetLevel(logLevel)

	logger.SetFormatter(&log.TextFormatter{
		TimestampFormat: "Jan 02 15:04:05.000",
		FullTimestamp:   true,
		ForceColors:     true,
	})

	logPath := CreateLogFolder()
	logFile := CreateLogFile(logPath, filename)
	mw := io.MultiWriter(os.Stdout, logFile)
	logger.SetOutput(mw)

	return logger.WithField("logger", prefix)
}

func CreateLogFolder() (fullPath string) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Cannot stat current working directory.")
	}

	fullPath, err = files.CreateFolder(cwd, logFolder)
	if err != nil {
		log.Fatal("Cannot open log folder.")
	}
	return fullPath
}

func CreateLogFile(folder string, filename string) (file *os.File) {
	file, err := files.OpenFile(folder, filename+logExt)
	if err != nil {
		log.Fatal("Cannot open log file.")
	}
	return file
}
