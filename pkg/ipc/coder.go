package ipc

import (
	"encoding/json"
)

func Encode(msg *Message) ([]byte, error) {
	return json.Marshal(msg)
}

func Decode(data []byte, msg *Message) error {
	return json.Unmarshal(data, msg)
}

// 解码到目标结构体
func DecodeTypedPayload(payload any, target any) error {
	if target == nil || payload == nil {
		return nil
	}
	b, ok := payload.([]byte)
	if !ok {
		tmp, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		b = tmp
	}
	return json.Unmarshal(b, target)
}
