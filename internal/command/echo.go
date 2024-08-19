package command

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/slack-go/slack"
)

type EchoCommand struct {
}

func (e *EchoCommand) Execute(ctx context.Context, api *slack.Client, command slack.SlashCommand, w http.ResponseWriter) error {
	params := &slack.Msg{Text: command.Text}
	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	return nil
}

func sendResponse(w http.ResponseWriter, response string) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}
