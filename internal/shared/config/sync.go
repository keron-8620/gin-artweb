package config

type StorageSyncConfig struct {
	Enable     bool   `yaml:"enable"`
	RemoteHost string `yaml:"remote_host"`
	RemotePort uint16 `yaml:"remote_port"`
	RemoteUser string `yaml:"remote_user"`
	RemotePath string `yaml:"remote_path"`
}
