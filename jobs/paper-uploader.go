package jobs

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/vishhvaan/lab-bot/files"
	"github.com/vishhvaan/lab-bot/functions"
	"github.com/vishhvaan/lab-bot/slack"
)

const outputTimeout = 5
const paperFolder = "papers"

type paperUploaderJob struct {
	labJob
	downloadFolder string
}

func (pu *paperUploaderJob) init() {
	pu.labJob.init()

	filepath, err := files.CreateFolder(files.FindExeDir(), paperFolder)
	if err != nil {
		message := "Cannot create paper download folder"
		slack.Message(message)
		pu.logger.Error(message)
		pu.active = false
	} else {
		pu.downloadFolder = filepath + string(os.PathSeparator)
	}
}

func (pu *paperUploaderJob) commandProcessor(c slack.CommandInfo) {
	if pu.active {
		controllerActions := map[string]action{
			"pmid": pu.paperPMIDUploader,
		}
		if len(c.Fields) == 1 {
			pu.errorMsg(c.Fields, c.Channel, "Include a URL or PMID with your request")
		} else {
			k := functions.GetKeys(controllerActions)
			subcommand := strings.ToLower(c.Fields[1])
			if functions.Contains(k, subcommand) {
				f := controllerActions[subcommand]
				f(c)
			} else {
				pu.paperDOIUploader(c)
			}
		}
	} else {
		slack.PostMessage(c.Channel, "The "+pu.name+" is disabled")
	}
}

func (pu *paperUploaderJob) errorMsg(fields []string, channel string, message string) {
	go pu.logger.WithField("fields", fields).Warn(message)
	slack.PostMessage(channel, message)
}

func (pu *paperUploaderJob) paperDOIUploader(c slack.CommandInfo) {
	paperURL := c.Fields[1]
	paperURL = paperURL[1 : len(paperURL)-1]
	url, err := url.ParseRequestURI(paperURL)

	if err != nil {
		pu.errorMsg(c.Fields, c.Channel, "Invalid URL")
	} else {
		command := fmt.Sprintf("scidownl download --doi \"%s\" --out %s", url.String(), pu.downloadFolder)
		output, err := slack.CommandStreamer(command, "err", c.Channel, outputTimeout)
		if err == nil {
			lastLine := output[len(output)-1]
			if strings.Contains(lastLine, "Successful") {
				i := strings.Index(lastLine, ": ")
				pdfPath := lastLine[i+2:]
				pu.logger.WithField("path", pdfPath).Info("Uploading File")
				slack.UploadFile(c.Channel, pdfPath, "")
				pu.logger.WithField("path", pdfPath).Info("Deleting File")
				files.DeleteFile(pdfPath)
			} else {
				pu.logger.Warn("Download not successful from scidownl")
				pu.errorMsg(c.Fields, c.Channel, "Could not upload paper")
			}
		} else {
			pu.logger.Error("Could not stream command")
			pu.errorMsg(c.Fields, c.Channel, "Could not upload paper")
		}
	}

}

func (pu *paperUploaderJob) paperPMIDUploader(c slack.CommandInfo) {

}
