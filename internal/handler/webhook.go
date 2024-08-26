package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/diabolusgx/snack-track/internal/env"
	"github.com/diabolusgx/snack-track/internal/models"
	"github.com/diabolusgx/snack-track/internal/util"
	"github.com/diabolusgx/snack-track/pkg/mongo"
	"github.com/slack-go/slack"
)

func RegisterWebhookHandler(api *slack.Client) {
	ctx := context.Background()

	http.HandleFunc("/webhook/order-update", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight (OPTIONS) request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// panic recovery
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
				log.Printf("[OrderUpdate] Recovered from panic: %v\n", r)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		// TODO: validate if request is actually coming from SnackTrack browser extension.

		buf, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("[OrderUpdate] Failed to read request body: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// decode the request body into model.OrderUpdate
		var orderUpdate *models.OrderUpdate
		if err := json.Unmarshal(buf, &orderUpdate); err != nil {
			log.Printf("[OrderUpdate] Failed to unmarshal request body: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		slackId := orderUpdate.SlackId
		if slackId == "" {
			log.Printf("[OrderUpdate] SlackId is empty\n")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// verify the slackId hash
		userId, ok, err := util.GetSlackIdFromHash(slackId)
		if err != nil {
			log.Printf("[OrderUpdate] Failed to verify slackId hash: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			log.Printf("[OrderUpdate] SlackId hash verification failed\n")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var user *models.User
		filters := mongo.Filters{
			{
				Key:      "user_id",
				Value:    userId,
				Type:     mongo.STRING,
				Operator: mongo.EQUAL,
			},
		}
		err = env.MongoClient().GetOne(ctx, env.MongoUsersCollectionName, filters, nil, &user)
		if err != nil {
			log.Printf("[OrderUpdate] Failed to get user: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		zOrder := orderUpdate.Order
		msg := fmt.Sprintf(
			"<@%s>'s order (`%d`) from %s is *%s* %s",
			userId,
			zOrder.OrderId,
			zOrder.ResInfo.Name,
			zOrder.DeliveryDetails.DeliveryLabel,
			zOrder.DeliveryDetails.DeliveryMessage,
		)
		_, _, _, err = api.SendMessage(user.ChannelId, slack.MsgOptionText(msg, false))
		if err != nil {
			log.Printf("[OrderUpdate] Failed to send message: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/webhook/user-settings", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight (OPTIONS) request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// panic recovery
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
				log.Printf("[UserSettings] Recovered from panic: %v\n", r)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		buf, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("[UserSettings] Failed to read request body: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// decode the request body into model.UpdateUserSettings
		var updateUserSettings *models.UpdateUserSettings
		if err := json.Unmarshal(buf, &updateUserSettings); err != nil {
			log.Printf("[UserSettings] Failed to unmarshal request body: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// verify the slackId hash
		slackId, ok, err := util.GetSlackIdFromHash(updateUserSettings.SlackId)
		if err != nil {
			log.Printf("[UserSettings] Failed to verify slackId hash: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			log.Printf("[UserSettings] SlackId hash verification failed\n")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if len(updateUserSettings.StartTime) != len(updateUserSettings.EndTime) {
			log.Printf("[UserSettings] Invalid start/end time\n")
			w.WriteHeader(http.StatusBadRequest)
		}

		// generate schedule
		var schedule []*models.Schedule
		for i := 0; i < len(updateUserSettings.StartTime); i++ {
			schedule = append(schedule, &models.Schedule{
				From: updateUserSettings.StartTime[i],
				To:   updateUserSettings.EndTime[i],
			})
		}

		var user *models.User
		filters := mongo.Filters{
			{
				Key:      "user_id",
				Value:    slackId,
				Type:     mongo.STRING,
				Operator: mongo.EQUAL,
			},
		}
		updates := mongo.Updates{
			{
				Key:            "schedule",
				Value:          schedule,
				Type:           mongo.STRING,
				UpdateOperator: mongo.SET,
			},
			{
				Key:            "address_ids",
				Value:          updateUserSettings.AddressIds,
				Type:           mongo.STRING_ARRAY,
				UpdateOperator: mongo.SET,
			},
		}
		err = env.MongoClient().FindOneAndUpdate(ctx, env.MongoUsersCollectionName, filters, updates, &user)
		if err == mongo.NoItemFound {
			err = env.MongoClient().Insert(ctx, env.MongoUsersCollectionName, &models.User{
				UserId:     slackId,
				Schedule:   schedule,
				AddressIds: updateUserSettings.AddressIds,
			})
		}
		if err != nil {
			log.Printf("[UserSettings] Failed to get user: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		msg := fmt.Sprintf("Your settings have been updated.\n\n%s", util.GetSlackMsgForSettings(user))
		_, _, _, err = api.SendMessage(slackId, slack.MsgOptionText(msg, false))
		if err != nil {
			log.Printf("[UserSettings] Failed to send message: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte("OK"))
	})

	fmt.Println("[INFO] Webhook handler registered")
}
