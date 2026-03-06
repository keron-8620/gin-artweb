package config

// UploadConfig 上传配置
type UploadConfig struct {
	MaxPkgSize    int `yaml:"max_pkg_size"`    // 最大上传程序包大小(MB)
	MaxScriptSize int `yaml:"max_script_size"` // 脚本最大上传大小(MB)
	MaxConfSize   int `yaml:"max_conf_size"`   // 配置文件最大上传大小(MB)
}
