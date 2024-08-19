package command

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/slack-go/slack"
)

type CommandExecutor interface {
	Execute(ctx context.Context, api *slack.Client, command slack.SlashCommand, w http.ResponseWriter) error
}

func NewCommandExecutor(s slack.SlashCommand) (CommandExecutor, error) {
	switch s.Command {
	case "/echo":
		return &EchoCommand{}, nil
	case "/track":
		return &TrackCommand{}, nil
	default:
		return nil, fmt.Errorf("unsupported command: %s", s.Command)
	}
}

// parseParams function takes an input string and an output struct (as an interface{}),
// extracts parameters, and decodes them into the struct.
func parseParams(input string, output interface{}) error {
	// Define a regex to find all parameters starting with '--'
	re := regexp.MustCompile(`--(\w+)=([^\s]+)`)
	matches := re.FindAllStringSubmatch(input, -1)

	// Create an empty map to hold the parameters
	paramsMap := make(map[string]interface{})

	// Iterate through the matches and store them in the map
	for _, match := range matches {
		key := strings.ToLower(match[1])
		value := match[2]

		// Store in the map
		paramsMap[key] = value
	}

	// Decode the map into the output struct
	err := mapstructure.Decode(paramsMap, &output)
	if err != nil {
		return err
	}

	return nil
}
