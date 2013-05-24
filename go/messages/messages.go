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

// The overloading of NUL is kind of unfortunate here, but ruby-land doesn't do
// the message-oriented thing so it's not a big deal.
func CreateCommandAndArgumentsMessage(command string, pid int, args []string) string {
	return "Q:" + command + ":" + strconv.Itoa(pid) + ":" + strings.Join(args, "\000")
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

func ParseClientCommandRequestMessage(msg string) (string, int, string, error) {
	parts := strings.SplitN(msg, ":", 4)
	if parts[0] != "Q" {
		return "", -1, "", errors.New("Wrong message type! Expected ClientCommandRequestMessage, got: " + msg)
	}

	command := parts[1]
	arguments := parts[3]
	pid, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", -1, "", errors.New("Expected pid, but none received: " + msg)
	}

	return command, pid, arguments, nil
}

func CreatePidAndArgumentsMessage(pid int, arguments string) string {
	return strconv.Itoa(pid) + ":" + arguments
}
