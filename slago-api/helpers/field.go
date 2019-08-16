// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).

package helpers

const (
	LevelFieldKey     = "level"
	TimestampFieldKey = "time"
	MessageFieldKey   = "message"
)

func getAndRemove(key string, event map[string]interface{}) string {
	field, ok := event[key]
	if !ok {
		return ""
	}
	delete(event, key)

	return field.(string)
}
