package event

import (
	"context"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func Handle(ctx context.Context, api *slack.Client, ev *slackevents.AppMentionEvent) {
	helpCommandContent := "Supported commands:\n" +
		"  - `help`: Display this help message\n" +
		"  - `echo <text>`: Echo back the text\n"

	ev.Text = strings.TrimSpace(ev.Text)
	commandArgs := strings.Split(ev.Text, " ")
	commandArgs = commandArgs[1:]
	if len(commandArgs) < 2 && BotCommand(commandArgs[0]) != HelpCommand {
		api.PostMessage(ev.Channel, slack.MsgOptionText(fmt.Sprintf("Invalid command, atleast 2 arguments are needed.\n %s", helpCommandContent), false))
		return
	}

	subCommand := BotCommand(commandArgs[0])
	switch subCommand {
	case HelpCommand:
		_, _, err := api.PostMessage(ev.Channel, slack.MsgOptionText(helpCommandContent, false))
		if err != nil {
			fmt.Println("Failed to post help message: ", err)
		}
	case EchoCommand:
		_, _, err := api.PostMessage(ev.Channel, slack.MsgOptionText(strings.Join(commandArgs[1:], " "), false))
		if err != nil {
			fmt.Println("Failed to post echo message: ", err)
		}
	default:
		err := api.AddReaction("x", slack.ItemRef{Channel: ev.Channel, Timestamp: ev.TimeStamp})
		if err != nil {
			fmt.Println("Failed to add reaction to message: ", err)
		}
		fmt.Println("Invalid command: ", subCommand)
	}
}
