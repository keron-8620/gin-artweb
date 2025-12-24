package config

import "time"

// DBConf 数据库配置结构体，用于配置数据库连接参数
type DBConf struct {
	Type            string `yaml:"type" json:"type"`                             // 数据库类型，支持 mysql, postgres, sqlite, sqlserver, opengauss
	Dns             string `yaml:"dns" json:"dns"`                               // 数据库连接字符串
	MaxIdleConns    int    `yaml:"max_idle_conns" json:"max_idle_conns"`         // 最大空闲连接数
	MaxOpenConns    int    `yaml:"max_open_conns" json:"max_open_conns"`         // 最大打开连接数
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`   // 连接最大生命周期(秒)
	ConnMaxIdleTime int    `yaml:"conn_max_idle_time" json:"conn_max_idle_time"` // 连接最大空闲时间(秒)
	LogSQL          bool   `yaml:"log_sql" json:"log_sql"`                       // 是否打印SQL
	ReadTimeout     int    `yaml:"read_timeout" json:"read_timeout"`             // 查询单条数据超时
	WriteTimeout    int    `yaml:"write_timeout" json:"write_timeout"`           // 写操作超时
	ListTimeout     int    `yaml:"list_timeout" json:"list_timeout"`             // 查询列表超时
}

// DBTimeout 数据库操作超时参数
type DBTimeout struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	ListTimeout  time.Duration
}
