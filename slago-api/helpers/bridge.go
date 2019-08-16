// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).

package helpers

import (
	"github.com/json-iterator/go"
	"gitlab.com/anbillon/slago/slago-api"
)

// BrigeWrite writes data from bridge to slago logger.
func BrigeWrite(bridge slago.Bridge, p []byte) error {
	var event map[string]interface{}
	if err := jsoniter.Unmarshal(p, &event); err != nil {
		return err
	}

	lvl := getAndRemove(LevelFieldKey, event)
	msg := getAndRemove(MessageFieldKey, event)
	delete(event, TimestampFieldKey)

	record := slago.Logger().Level(bridge.ParseLevel(lvl))
	for k, v := range event {
		record.Interface(k, v)
	}
	record.Msg(msg)

	return nil
}
