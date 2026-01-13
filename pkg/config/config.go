package config

import "time"

// Config 全局配置结构
type Config struct {
	Server   ServerConfig            `mapstructure:"server"`
	Adapters map[string]AdapterConfig `mapstructure:"adapters"`
	Router   RouterConfig            `mapstructure:"router"`
	Logging  LoggingConfig           `mapstructure:"logging"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"` // debug, release
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// AdapterConfig 适配器配置
type AdapterConfig struct {
	APIKey  string        `mapstructure:"api_key"`
	BaseURL string        `mapstructure:"base_url"`
	Timeout time.Duration `mapstructure:"timeout"`
	Models  []string      `mapstructure:"models"`
}

// RouterConfig 路由配置
type RouterConfig struct {
	DefaultStrategy string                 `mapstructure:"default_strategy"`
	Strategies      map[string]interface{} `mapstructure:"strategies"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `mapstructure:"level"`       // debug, info, warn, error
	Encoding   string `mapstructure:"encoding"`    // json, console
	Output     string `mapstructure:"output"`      // stdout, file
	FilePath   string `mapstructure:"file_path"`   // 日志文件路径
	MaxSize    int    `mapstructure:"max_size"`    // MB
	MaxBackups int    `mapstructure:"max_backups"` // 最大备份数
	MaxAge     int    `mapstructure:"max_age"`     // 保留天数
}
