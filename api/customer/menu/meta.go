package menu

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap/zapcore"
)

type MetaSchemas struct {
	// 标题
	Title string `json:"title" example:"用户管理"`
	// 图标
	Icon string `json:"icon" example:"icon"`
}

func (m *MetaSchemas) Json() string {
	jd, _ := json.Marshal(m)
	return string(jd)
}

func (m *MetaSchemas) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("title", m.Title)
	enc.AddString("icon", m.Icon)
	return nil
}

func NewMetaSchemas(ms string) (*MetaSchemas, error) {
	if ms == "" {
		return nil, fmt.Errorf("meta is empty")
	}
	var meta MetaSchemas
	err := json.Unmarshal([]byte(ms), &meta)
	if err != nil {
		return nil, fmt.Errorf("解析 MetaSchemas 失败: %w", err)
	}
	return &meta, nil
}
