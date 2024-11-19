package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

const MagicKey string = "_magic_key_"

// 以 JSON 格式保存
//
// 读取时不解析 仅将字符串存在 MagicKey 键下
type Header map[string]string

func (Header) GormDataType() string {
	return "TEXT"
}

func (h *Header) Scan(src any) error {
	if src == nil {
		*h = make(Header)
		return nil
	}
	switch src := src.(type) {
	case []byte:
		*h = Header{MagicKey: string(src)}
	case string:
		*h = Header{MagicKey: src}
	default:
		return fmt.Errorf("model: failed to unmarshal Header value: %v", src)
	}
	return nil
}

func (h Header) Value() (driver.Value, error) {
	if len(h) == 0 {
		return "{}", nil
	}
	if v, ok := h[MagicKey]; ok {
		return v, nil
	}
	b, err := json.Marshal(map[string]string(h))
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (h Header) MarshalJSON() ([]byte, error) {
	if h == nil {
		return nil, nil
	}
	if len(h) == 0 {
		return []byte("{}"), nil
	}
	if v, ok := h[MagicKey]; ok {
		return []byte(v), nil
	}
	return json.Marshal(map[string]string(h))
}

func (h Header) String() string {
	if v, err := h.Value(); err == nil {
		return v.(string)
	}
	return ""
}
