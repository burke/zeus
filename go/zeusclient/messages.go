package zeusclient

import (
	"encoding/json"
	"strconv"
)

func CreateCommandAndArgumentsMessage(command string, pid int, args []string) string {
	encoded, _ := json.Marshal(args)
	return "Q:" + command + ":" + strconv.Itoa(pid) + ":" + string(encoded)
}
