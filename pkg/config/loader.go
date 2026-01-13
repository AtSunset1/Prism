package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var (
	// GlobalConfig 全局配置实例
	GlobalConfig *Config
)

// Load 加载配置文件
// configPath: 配置文件路径，如 "configs/config.yaml"
// 返回: Config实例和错误信息
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 1. 设置配置文件路径
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 默认配置文件搜索路径
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath("../configs")
		v.AddConfigPath(".")
	}

	// 2. 设置默认值（最低优先级）
	setDefaults(v)

	// 3. 显式绑定环境变量（更清晰，易于维护）
	bindEnvVars(v)

	// 4. 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，只使用环境变量和默认值
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintf(os.Stderr, "Warning: Config file not found, using environment variables and defaults\n")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		fmt.Printf("Using config file: %s\n", v.ConfigFileUsed())
	}

	// 5. 解析到Config结构体
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 6. 验证配置
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// 保存到全局变量
	GlobalConfig = cfg

	return cfg, nil
}

// setDefaults 设置默认配置
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.encoding", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("logging.file_path", "./logs/prism.log")
	v.SetDefault("logging.max_size", 100)
	v.SetDefault("logging.max_backups", 10)
	v.SetDefault("logging.max_age", 30)

	// Router defaults
	v.SetDefault("router.default_strategy", "simple")
}

// bindEnvVars 显式绑定环境变量
// 这种方式比 AutomaticEnv 更清晰，易于维护和理解
func bindEnvVars(v *viper.Viper) {
	// Server 配置绑定
	v.BindEnv("server.host", "SERVER_HOST")
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.mode", "SERVER_MODE")
	v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")

	// Logging 配置绑定
	v.BindEnv("logging.level", "LOG_LEVEL")
	v.BindEnv("logging.encoding", "LOG_ENCODING")
	v.BindEnv("logging.output", "LOG_OUTPUT")
	v.BindEnv("logging.file_path", "LOG_FILE_PATH")

	// Router 配置绑定
	v.BindEnv("router.default_strategy", "ROUTER_STRATEGY")

	// Adapter 配置绑定（API密钥）
	// GLM 适配器
	v.BindEnv("adapters.glm.api_key", "GLM_API_KEY")
	v.BindEnv("adapters.glm.base_url", "GLM_BASE_URL")
	v.BindEnv("adapters.glm.timeout", "GLM_TIMEOUT")

	// 豆包适配器（预留）
	v.BindEnv("adapters.doubao.api_key", "DOUBAO_API_KEY")
	v.BindEnv("adapters.doubao.base_url", "DOUBAO_BASE_URL")
	v.BindEnv("adapters.doubao.timeout", "DOUBAO_TIMEOUT")

	// 文心一言适配器（预留）
	v.BindEnv("adapters.wenxin.api_key", "WENXIN_API_KEY")
	v.BindEnv("adapters.wenxin.base_url", "WENXIN_BASE_URL")
	v.BindEnv("adapters.wenxin.timeout", "WENXIN_TIMEOUT")
}

// validate 验证配置
func validate(cfg *Config) error {
	// 验证服务器配置
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	if cfg.Server.Mode != "debug" && cfg.Server.Mode != "release" {
		return fmt.Errorf("invalid server mode: %s (must be 'debug' or 'release')", cfg.Server.Mode)
	}

	// 验证适配器配置
	if len(cfg.Adapters) == 0 {
		return fmt.Errorf("no adapters configured")
	}

	for name, adapter := range cfg.Adapters {
		if adapter.APIKey == "" {
			return fmt.Errorf("adapter '%s' missing API key", name)
		}
		if adapter.BaseURL == "" {
			return fmt.Errorf("adapter '%s' missing base URL", name)
		}
		if len(adapter.Models) == 0 {
			return fmt.Errorf("adapter '%s' has no models configured", name)
		}
	}

	// 验证日志配置
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Logging.Level] {
		return fmt.Errorf("invalid logging level: %s", cfg.Logging.Level)
	}

	return nil
}

// GetConfig 获取全局配置实例
func GetConfig() *Config {
	return GlobalConfig
}
