package env

import (
	"context"

	"github.com/diabolusgx/snack-track/pkg/mongo"
	"github.com/joho/godotenv"
)

type ctxKey string

const (
	// context keys
	envKey ctxKey = "env"

	// environment variables
	SlackBotToken      = "SLACK_BOT_TOKEN"
	SlackSigningSecret = "SLACK_SIGNING_SECRET"
	MongoConnectionURI = "MONGO_CONNECTION_URI"

	// global constants
	MongoDatabaseName        = "snack-track"
	MongoUsersCollectionName = "users"
)

type Env struct {
	vars map[string]string

	mongoClient *mongo.MongoDB
}

var env *Env

func init() {
	vars, err := godotenv.Read(".env")
	if err != nil {
		panic(err)
	}
	env = &Env{vars: vars}
}

func GetEnv() *Env {
	return env
}

func GetEnvFromContext(ctx context.Context) *Env {
	return ctx.Value(envKey).(*Env)
}

func GetContextWithEnv(ctx context.Context, env *Env) context.Context {
	return context.WithValue(ctx, envKey, env)
}

func WithMongoClient(client *mongo.MongoDB) {
	env.mongoClient = client
}

func MongoClient() *mongo.MongoDB {
	return env.mongoClient
}

func GetParam(key string) (string, bool) {
	val, found := env.vars[key]
	return val, found
}
