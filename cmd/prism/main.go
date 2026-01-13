package main

import (
	"fmt"
	"log"

	"github.com/AtSunset1/prism/internal/adapter"
	"github.com/AtSunset1/prism/internal/adapter/glm"
	"github.com/AtSunset1/prism/internal/handler"
	"github.com/AtSunset1/prism/internal/router"
	"github.com/AtSunset1/prism/pkg/config"
)

func main() {
	// 1. 加载配置
	cfg := loadConfig()

	// 2. 初始化适配器和处理器
	chatHandler := initHandlers(cfg)

	// 3. 设置路由
	r := router.SetupRouter(chatHandler)

	// 4. 启动服务器
	startServer(r, cfg)
}

// loadConfig 加载配置文件
// 返回：*config.Config 配置实例
func loadConfig() *config.Config {
	log.Println("========================================")
	log.Println("📋 加载配置文件...")

	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	log.Println("✓ 配置加载成功")
	log.Printf("✓ 服务器模式: %s", cfg.Server.Mode)
	log.Printf("✓ 服务器端口: %d", cfg.Server.Port)
	log.Printf("✓ 日志级别: %s", cfg.Logging.Level)
	log.Printf("✓ 已配置适配器: %d 个", len(cfg.Adapters))
	log.Println("========================================")

	return cfg
}

// initHandlers 初始化适配器和处理器
// 参数：
//   - cfg: 配置实例
// 返回：
//   - *handler.ChatHandler: 聊天处理器
func initHandlers(cfg *config.Config) *handler.ChatHandler {
	log.Println("🔧 初始化适配器...")

	// 创建适配器管理器
	manager := adapter.NewAdapterManager()

	// 遍历配置，动态注册适配器
	for adapterName, adapterCfg := range cfg.Adapters {
		log.Printf("  └─ 初始化适配器: %s", adapterName)

		// 根据适配器类型创建实例
		var adp adapter.ModelAdapter
		switch adapterName {
		case "glm":
			// 创建GLM适配器
			adp = glm.NewGLMAdapter(adapterCfg.APIKey)
			log.Printf("     ✓ GLM适配器创建成功 (API Key: %s...)", maskAPIKey(adapterCfg.APIKey))

			// 为每个模型注册适配器
			for _, modelName := range adapterCfg.Models {
				if err := manager.Register(modelName, adp); err != nil {
					log.Fatalf("❌ 注册模型 %s 失败: %v", modelName, err)
				}
				log.Printf("     ✓ 模型 %s 注册成功", modelName)
			}

		default:
			log.Printf("     ⚠️  跳过未实现的适配器: %s", adapterName)
		}
	}

	log.Println("✓ 适配器管理器初始化成功")
	log.Printf("✓ 已注册模型: %v", manager.ListModels())

	// 创建ChatHandler
	chatHandler := handler.NewChatHandler(manager)
	log.Println("✓ ChatHandler初始化成功")
	log.Println("========================================")

	return chatHandler
}

// startServer 启动HTTP服务器
// 参数：
//   - r: Gin路由器
//   - cfg: 配置实例
func startServer(r interface{ Run(addr ...string) error }, cfg *config.Config) {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	log.Println("🚀 Prism AI Gateway 启动成功")
	log.Printf("📍 监听地址: http://localhost:%d", cfg.Server.Port)
	log.Printf("🔧 运行模式: %s", cfg.Server.Mode)
	log.Println("📚 可用接口:")
	log.Println("   - GET  /              欢迎页面")
	log.Println("   - GET  /health        健康检查")
	log.Println("   - POST /v1/chat/completions  聊天补全")
	log.Println("========================================")

	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ 启动失败: %v", err)
	}
}

// maskAPIKey 隐藏API密钥（只显示前8位）
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:8] + "..."
}