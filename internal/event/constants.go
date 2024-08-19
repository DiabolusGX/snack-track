package event

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// BotCommand is a type for supported bot commands
type BotCommand string

const (
	// HelpCommand is the help command
	HelpCommand BotCommand = "help"
	// EchoCommand is the echo command
	EchoCommand BotCommand = "echo"
)

type EventHandler interface {
	Handle(ctx context.Context, api *slack.Client, ev *slackevents.AppMentionEvent)
}
