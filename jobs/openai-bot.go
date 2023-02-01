package jobs

import (
	"context"
	"strings"
	"time"

	gogpt "github.com/sashabaranov/go-gpt3"

	"github.com/vishhvaan/lab-bot/config"
	"github.com/vishhvaan/lab-bot/functions"
	"github.com/vishhvaan/lab-bot/slack"
)

type openAIBot struct {
	labJob
	gptClient        *gogpt.Client
	gptContext       context.Context
	defaultTimeout   time.Duration
	model            string
	maxTokens        int
	Temperature      float32
	TopP             float32
	FrequencyPenalty float32
	PresencePenalty  float32
}

func (b *openAIBot) init() {
	b.labJob.init()

	if !functions.Contains(functions.GetKeys(config.Secrets), "openai-api-key") {
		b.logger.Error("OpenAI API Key not found in the secrets file (key is openai-api-key)")
		go slack.Message("OpenAI API key not found. Disabling response bot.")
		b.active = false
		return
	}

	b.gptClient = gogpt.NewClient(config.Secrets["openai-api-key"])
	b.gptContext = context.Background()
	b.defaultTimeout = 10 * time.Second

	b.model = "text-davinci-003"
	b.maxTokens = 1000
	b.Temperature = 0.5
	b.TopP = 0.3
	b.FrequencyPenalty = 0.5
	b.PresencePenalty = 0

	m := "The OpenAI chat bot has been loaded."
	go slack.Message(m)
	b.logger.Info(m)
}

func (b *openAIBot) commandProcessor(c slack.CommandInfo) {
	if b.active {
		controllerActions := map[string]action{
			"modify": b.modifyParameters,
		}
		if len(c.Fields) == 1 {
			slack.PostMessage(c.Channel, "No message detected")
		} else {
			k := functions.GetKeys(controllerActions)
			subcommand := strings.ToLower(c.Fields[1])
			if functions.Contains(k, subcommand) {
				f := controllerActions[subcommand]
				f(c)
			} else {
				b.sendCompletion(c)
			}
		}
	} else {
		slack.PostMessage(c.Channel, "The "+b.name+" is disabled")
	}
}

func (b *openAIBot) sendCompletion(c slack.CommandInfo) {
	prompt := strings.Join(c.Fields[1:], " ")
	req := gogpt.CompletionRequest{
		Model:     b.model,
		MaxTokens: b.maxTokens,
		Prompt:    prompt,
	}

	deadline := time.Now().Add(b.defaultTimeout)
	cont, cancel := context.WithDeadline(b.gptContext, deadline)
	defer cancel()

	resp, err := b.gptClient.CreateCompletion(cont, req)
	if err != nil {
		go b.logger.WithField("prompt", prompt).WithError(err).Warn("Could not find response for the prompt")
		slack.PostMessage(c.Channel, "Could not find response for the prompt. "+err.Error())
		return
	}

	m := resp.Choices[0].Text
	go slack.PostMessage(c.Channel, m)
	b.logger.WithField("prompt", prompt).Info(m)
}

func (b *openAIBot) modifyParameters(c slack.CommandInfo) {

}
