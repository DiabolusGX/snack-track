package command

import (
	"context"
	"log"
	"net/http"

	"github.com/diabolusgx/snack-track/internal/env"
	"github.com/diabolusgx/snack-track/internal/models"
	"github.com/diabolusgx/snack-track/internal/util"
	"github.com/diabolusgx/snack-track/pkg/mongo"
	"github.com/slack-go/slack"
)

type StSettings struct {
}

func (t *StSettings) Execute(ctx context.Context, api *slack.Client, command *slack.SlashCommand, w http.ResponseWriter) error {
	var user *models.User
	filters := mongo.Filters{
		{
			Key:      "user_id",
			Value:    command.UserID,
			Type:     mongo.STRING,
			Operator: mongo.EQUAL,
		},
	}
	err := env.MongoClient().GetOne(ctx, env.MongoUsersCollectionName, filters, nil, &user)
	if err == mongo.NoItemFound {
		sendResponse(w, "You have not set up your SnackTrack settings yet.\nPlease use `/st-channel`, `/st-token` and Snack Track extension to get started.")
		return nil
	}
	if err != nil {
		log.Printf("[StSettings] Failed to get user: %v\n", err)
		return err
	}

	sendResponse(w, util.GetSlackMsgForSettings(user))
	return nil
}
