package jobs

import (
	"context"
	"strings"
	"time"

	gogpt "github.com/sashabaranov/go-gpt3"

	"github.com/vishhvaan/lab-bot/slack"
)

type openAIBot struct {
	labJob
	gptClient      *gogpt.Client
	gptContext     context.Context
	model          string
	defaultTimeout time.Duration
}

func (b *openAIBot) init() {
	b.labJob.init()

	b.gptClient = gogpt.NewClient("your token")
	b.gptContext = context.Background()
	b.defaultTimeout = 10 * time.Second
}

func (b *openAIBot) sendCompletion(c slack.CommandInfo) {
	prompt := strings.Join(c.Fields[2:], " ")
	req := gogpt.CompletionRequest{
		Model:     b.model,
		MaxTokens: 5,
		Prompt:    prompt,
	}

	deadline := time.Now().Add(b.defaultTimeout)
	cont, cancel := context.WithDeadline(b.gptContext, deadline)
	defer cancel()

	resp, err := b.gptClient.CreateCompletion(cont, req)
	if err != nil {
		go b.logger.WithField("prompt", prompt).WithError(err).Warn("Could not find response for the prompt")
		slack.PostMessage(c.Channel, "Could not find response for the prompt")
		return
	}

	m := resp.Choices[0].Text
	go b.logger.WithField("prompt", prompt).Info(m)
	slack.PostMessage(c.Channel, m)
}
