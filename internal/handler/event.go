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
	"github.com/diabolusgx/snack-track/internal/event"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func RegisterEventAPIHandler(api *slack.Client) {
	signingSecret, found := env.GetParam(env.SlackSigningSecret)
	if !found {
		panic("SLACK_SIGNING_SECRET is not set")
	}

	http.HandleFunc("/slack/event", func(w http.ResponseWriter, r *http.Request) {
		// panic recovery
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
				log.Printf("[SlackEventHandler] Recovered from panic: %v\n", r)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			fmt.Println("[SlackEventHandler] Failed to create secrets verifier:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err := sv.Write(body); err != nil {
			fmt.Println("[SlackEventHandler] Failed to write body:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			fmt.Println("[SlackEventHandler] Failed to ensure signature:", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			fmt.Println("[SlackEventHandler] Failed to parse event:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				fmt.Println("[SlackEventHandler] Failed to unmarshal challenge response:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
			return
		}

		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				ctx := context.Background()
				event.Handle(ctx, api, ev)
				return
			}
		}
		fmt.Println("[INFO] Unhandled event:", eventsAPIEvent.Type)
	})

	fmt.Println("[INFO] Event API handler registered")
}
