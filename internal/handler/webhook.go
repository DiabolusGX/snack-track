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
	"github.com/diabolusgx/snack-track/pkg/mongo"
	"github.com/slack-go/slack"
)

func RegisterWebhookHandler(api *slack.Client) {
	ctx := context.Background()

	http.HandleFunc("/webhook/order-update", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
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

		// TODO: figure out user <> order mapping
		userId := "test-user-id"

		var user models.User
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

	fmt.Println("[INFO] Webhook handler registered")
}
