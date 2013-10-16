package messages

import (
	"errors"
	"strconv"
	"strings"
)

func ParsePidMessage(msg string) (int, string, error) {
	parts := strings.SplitN(msg, ":", 3)
	if parts[0] != "P" {
		return -1, "", errors.New("Wrong message type! Expected PidMessage, got: " + msg)
	}

	identifier := parts[2]
	pid, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1, "", err
	}

	return pid, identifier, nil
}

func CreateCommandAndArgumentsMessage(args []string, pid int) string {
	return "T:" + strconv.Itoa(len(args)-1) + ":" + strconv.Itoa(pid) + ":" + args[0]
}

func ParseFeatureMessage(msg string) (string, error) {
	parts := strings.SplitN(msg, ":", 2)
	if parts[0] != "F" {
		return "", errors.New("Wrong message type! Expected FeatureMessage, got: " + msg)
	}
	return strings.TrimSpace(parts[1]), nil
}

func ParseActionResponseMessage(msg string) (string, error) {
	parts := strings.SplitN(msg, ":", 2)
	if parts[0] != "R" {
		return "", errors.New("Wrong message type! Expected ActionResponseMessage, got: " + msg)
	}
	return parts[1], nil
}

func CreateSpawnSlaveMessage(identifier string) string {
	return "S:" + identifier
}

func CreateSpawnCommandMessage(identifier string) string {
	return "C:" + identifier
}

func ParseClientCommandRequestMessage(msg string) (int, int, string, error) {
	parts := strings.SplitN(msg, ":", 4)
	if parts[0] != "T" {
		return -1, -1, "", errors.New("Wrong message type! Expected ClientCommandRequestMessage, got: " + msg)
	}

	argLength, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1, -1, "", errors.New("Expected argument count, but none received: " + msg)
	}
	pid, err := strconv.Atoi(parts[2])
	if err != nil {
		return argLength, -1, "", errors.New("Expected pid, but none received: " + msg)
	}

	return argLength, pid, parts[3], nil
}

func CreatePidAndArgumentsMessage(pid int, argCount int) string {
	return strconv.Itoa(pid) + ":" + strconv.Itoa(argCount)
}
