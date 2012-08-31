package zeusclient

import (
	"encoding/json"
)

func CreateCommandAndArgumentsMessage(command string, args []string) string {
	encoded, _ := json.Marshal(args)
	return "Q:" + command + ":" + string(encoded) + "\n"
}
