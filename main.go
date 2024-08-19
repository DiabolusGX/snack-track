package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/diabolusgx/snack-track/internal/env"
	"github.com/diabolusgx/snack-track/internal/handler"
	"github.com/diabolusgx/snack-track/pkg/mongo"
	"github.com/slack-go/slack"
)

func main() {
	botToken, found := env.GetParam(env.SlackBotToken)
	if !found {
		panic("SLACK_BOT_TOKEN is not set")
	}
	api := slack.New(botToken)

	connectionURI, found := env.GetParam(env.MongoConnectionURI)
	if !found {
		panic("MONGO_CONNECTION_URI is not set")
	}
	client := mongo.NewMongoDB(context.TODO(), env.MongoDatabaseName, connectionURI)
	env.WithMongoClient(client)

	handler.RegisterEventAPIHandler(api)
	handler.RegisterCommandAPIHandler(api)

	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":2929", nil)
}
