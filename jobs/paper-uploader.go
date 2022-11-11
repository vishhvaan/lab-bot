package jobs

import (
	"fmt"
	"net/url"
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
	_, err := url.ParseRequestURI(paperURL)

	if err != nil {
		pu.errorMsg(c.Fields, c.Channel, "Invalid URL")
	} else {
		command := fmt.Sprintf("scidownl download --doi %s", paperURL)
		output, err := slack.CommandStreamer(command, "err", c.Channel, outputTimeout)
		if err == nil {

		}
	}

}

func (pu *paperUploaderJob) paperPMIDUploader(c slack.CommandInfo) {

}
