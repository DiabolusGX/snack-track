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

type StChannel struct {
}

func (t *StChannel) Execute(ctx context.Context, api *slack.Client, command *slack.SlashCommand, w http.ResponseWriter) error {
	filters := mongo.Filters{
		{
			Key:      "user_id",
			Value:    command.UserID,
			Type:     mongo.STRING,
			Operator: mongo.EQUAL,
		},
	}
	updates := mongo.Updates{
		{
			Key:            "channel_id",
			Value:          command.ChannelID,
			Type:           mongo.STRING_ARRAY,
			UpdateOperator: mongo.SET,
		},
		{
			Key:            "team_domain",
			Value:          command.TeamDomain,
			Type:           mongo.STRING_ARRAY,
			UpdateOperator: mongo.SET,
		},
	}
	err := env.MongoClient().Upsert(ctx, env.MongoUsersCollectionName, filters, updates)
	if err != nil {
		log.Printf("[UserSettings] Failed to upsert user: %v\n", err)
		return err
	}

	var user *models.User
	filters = mongo.Filters{
		{
			Key:      "user_id",
			Value:    command.UserID,
			Type:     mongo.STRING,
			Operator: mongo.EQUAL,
		},
	}
	err = env.MongoClient().GetOne(ctx, env.MongoUsersCollectionName, filters, nil, &user)
	if err != nil {
		log.Printf("[StSettings] Failed to get user: %v\n", err)
		return err
	}

	sendResponse(w, "Your channel has been set to <#"+command.ChannelID+">"+"\n\n"+util.GetSlackMsgForSettings(user))
	return nil
}
