package config

import "time"

type APIConfig struct {
	TotalTimeout     time.Duration `yaml:"total_timeout"`
	BizCreateTimeout time.Duration `yaml:"biz_create_timeout"`
	BizUpdateTimeout time.Duration `yaml:"biz_update_timeout"`
	BizQueryTimeout  time.Duration `yaml:"biz_query_timeout"`
	BizDeleteTimeout time.Duration `yaml:"biz_delete_timeout"`
}
