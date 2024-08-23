package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/diabolusgx/snack-track/internal/env"
	"github.com/diabolusgx/snack-track/internal/models"
)

func GetSlackMsgForSettings(user *models.User) string {
	channelMsg := "But you need to set updates channel first. Please use `/set-channel` command in channel which you want to use for order updates.\n"
	if user.ChannelId != "" {
		channelMsg = "Your updates channel is set to <#" + user.ChannelId + ">\n"
	}

	addressMsg := "There's no address filter. Hence, orders on all addresses will be sent to your updates channel.\n"
	if len(user.AddressIds) > 0 {
		addressMsg = "Orders on the following addresses will be sent to your updates channel: `" + strings.Join(user.AddressIds, "`, ") + "`\n"
	}

	timeMsg := "You have not set any schedule. Hence, you will receive updates irrespective of time of placing order.\n"
	if len(user.Schedule) > 0 {
		timeMsg = "You will receive updates for orders between the following times:\n"
		for _, s := range user.Schedule {
			timeMsg += "- " + s.From + " to " + s.To + "\n"
		}
	}

	return "Here are your settings:\n" + channelMsg + addressMsg + timeMsg
}

func GetSlackIdFromHash(slackId string) (string, bool, error) {
	params := strings.Split(slackId, "#")
	if len(params) != 2 {
		return "", false, errors.New("invalid slackId format")
	}

	ok, err := verifyHash(params[0], params[1])
	return params[0], ok, err
}

func GetHashFromSlackId(slackId string) (string, error) {
	key, found := env.GetParam(env.SecretKey)
	if !found {
		return "", errors.New("secret key not found")
	}

	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(slackId))
	return slackId + "#" + hex.EncodeToString(mac.Sum(nil)), nil
}

func verifyHash(data, hash string) (bool, error) {
	key, found := env.GetParam(env.SecretKey)
	if !found {
		return false, errors.New("secret key not found")
	}

	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	expectedHash := hex.EncodeToString(mac.Sum(nil))
	return hash == expectedHash, nil
}
