package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/diabolusgx/snack-track/internal/command"
	"github.com/diabolusgx/snack-track/internal/env"
	"github.com/slack-go/slack"
)

func RegisterCommandAPIHandler(api *slack.Client) {
	signingSecret, found := env.GetParam(env.SlackSigningSecret)
	if !found {
		panic("SLACK_SIGNING_SECRET is not set")
	}

	http.HandleFunc("/slack/command", func(w http.ResponseWriter, r *http.Request) {
		// panic recovery
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
				log.Printf("[SlackCommandHandler] Recovered from panic: %v\n", r)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(io.TeeReader(r.Body, &verifier))
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err = verifier.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		executor, err := command.NewCommandExecutor(s)
		if err != nil {
			fmt.Println("[SlackCommandHandler] Failed to create command executor:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := context.Background()
		err = executor.Execute(ctx, api, &s, w)
		if err != nil {
			fmt.Println("[SlackCommandHandler] Failed to execute command:", s.Command, err)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf("unexpected error: %s", err.Error())))
		}
	})

	fmt.Println("[INFO] Command API handler registered")
}
