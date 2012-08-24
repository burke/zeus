package messages

import (
	"strings"
	"strconv"
	"errors"
)

func ParsePidMessage(msg string) (int, string, error) {
	parts := strings.SplitN(msg, ":", 3)
	if parts[0] != "P" {
		return -1, "", errors.New("Wrong message type!")
	}

	identifier := parts[2]
	pid, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1, "", err
	}

	return pid, identifier, nil
}

func CreateActionMessage(action string) (string) {
	return "A:" + action
}

func ParseActionResponseMessage(msg string) (string, error) {
	parts := strings.SplitN(msg, ":", 2)
	if parts[0] != "R" {
		return "", errors.New("Wrong message type!")
	}
	return parts[1], nil
}

func CreateSpawnSlaveMessage(identifier string) (string) {
	return "S:" + identifier
}

func CreateSpawnCommandMessage(identifier string) (string) {
	return "C:" + identifier
}

