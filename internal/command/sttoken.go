package command

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/diabolusgx/snack-track/internal/util"
	"github.com/slack-go/slack"
)

type StToken struct {
}

func (t *StToken) Execute(ctx context.Context, api *slack.Client, command *slack.SlashCommand, w http.ResponseWriter) error {
	hash, err := util.GetHashFromSlackId(command.UserID)
	if err != nil {
		return err
	}

	sendResponse(w, hash)

	_, _, _, err = api.SendMessage(command.UserID, slack.MsgOptionText(fmt.Sprintf("You *Snack Track* browser extension id is: %s", hash), false))
	if err != nil {
		log.Printf("[StToken] Failed to send message to user: %v", err)
	}
	return nil
}
