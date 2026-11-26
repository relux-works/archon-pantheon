package sessiondriver

import (
	"bytes"
	"encoding/json"
)

func DebugPayloadEnvelope(frame []byte) bool {
	if len(frame) == 0 || len(frame) > MaxFrame || !json.Valid(frame) || bytes.TrimSpace(frame)[0] != '{' || !uniqueKeys(frame) {
		return false
	}
	var e envelope
	return json.Unmarshal(frame, &e) == nil && e.Method != "" && len(e.Result) == 0 && len(e.Error) == 0
}
