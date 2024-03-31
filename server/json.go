package server

import (
	"encoding/json"
)

func unmarshalString(data string, obj interface{}) error {
	return unmarshalData([]byte(data), obj)
}

func unmarshalData(data []byte, obj interface{}) error {
	return json.Unmarshal(data, obj)
}
