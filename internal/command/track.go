package command

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/diabolusgx/snack-track/internal/env"
	"github.com/diabolusgx/snack-track/internal/models"
	"github.com/diabolusgx/snack-track/internal/shared"
	"github.com/diabolusgx/snack-track/pkg/mongo"
	"github.com/slack-go/slack"
)

type TrackCommand struct {
	From string `mapstructure:"from"`
	To   string `mapstructure:"to"`
}

func (t *TrackCommand) Execute(ctx context.Context, api *slack.Client, command *slack.SlashCommand, w http.ResponseWriter) error {
	err := parseParams(command.Text, &t)
	if err != nil {
		return err
	}

	invalidTimeMsg := "`--from` and `--to` params must contain valid time like `--from=09:00` and `--to=17:00`"
	if t.From == "" || t.To == "" {
		sendResponse(w, invalidTimeMsg)
		return nil
	}
	fromTime, err := time.Parse(shared.ScheduleTimeFormat, t.From)
	if err != nil {
		sendResponse(w, invalidTimeMsg)
		return nil
	}
	toTime, err := time.Parse(shared.ScheduleTimeFormat, t.To)
	if err != nil {
		sendResponse(w, invalidTimeMsg)
		return nil
	}

	if fromTime.After(toTime) {
		sendResponse(w, "`--from` time must be before `--to` time")
		return nil
	}
	fromStr := fromTime.Format(shared.ScheduleTimeFormat)
	toStr := toTime.Format(shared.ScheduleTimeFormat)

	newUser := false
	var user *models.User
	keyFilters := mongo.Filters{
		{
			Key:      "user_id",
			Value:    command.UserID,
			Operator: mongo.EQUAL,
			Type:     mongo.STRING,
		},
	}
	err = env.MongoClient().GetOne(ctx, env.MongoUsersCollectionName, keyFilters, nil, &user)
	if err == mongo.NoItemFound {
		newUser = true
		user = &models.User{
			UserId:     command.UserID,
			ChannelId:  command.ChannelID,
			TeamDomain: command.TeamDomain,
			Schedule: []*models.Schedule{
				{
					From: fromStr,
					To:   toStr,
				},
			},
		}
		err = env.MongoClient().Insert(ctx, env.MongoUsersCollectionName, user)
		if err != nil {
			log.Printf("[ERROR] failed to insert user to database, err: %s\n", err.Error())
			return err
		}
	}
	if err != nil {
		log.Printf("[ERROR] failed to get user from database, err: %s\n", err.Error())
		return err
	}

	if !newUser {
		updates := mongo.Updates{
			{
				Key:            "schedule",
				Value:          []*models.Schedule{{From: fromStr, To: toStr}},
				UpdateOperator: mongo.PUSH,
			},
		}
		err = env.MongoClient().FindOneAndUpdate(ctx, env.MongoUsersCollectionName, keyFilters, updates, &user)
		if err != nil {
			log.Printf("[ERROR] failed to update user to database, err: %s\n", err.Error())
			return err
		}
	}

	strBuilder := &strings.Builder{}
	strBuilder.WriteString("We'll track your delivery orders: \n")
	for _, schedule := range user.Schedule {
		strBuilder.WriteString("From ")
		strBuilder.WriteString(schedule.From)
		strBuilder.WriteString(" — To ")
		strBuilder.WriteString(schedule.To)
		strBuilder.WriteString("\n")
	}
	sendResponse(w, strBuilder.String())
	return nil
}
