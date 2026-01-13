package main

import (
	"log"
	"os"

	"github.com/AtSunset1/prism/internal/adapter"
	"github.com/AtSunset1/prism/internal/adapter/glm"
	"github.com/AtSunset1/prism/internal/handler"
	"github.com/AtSunset1/prism/internal/router"
)

// TestAPIKey 测试用API密钥（仅用于开发测试）
// 生产环境应使用环境变量 GLM_API_KEY
const TestAPIKey = ""

func main() {
	// 1. 初始化配置
	apiKey := initConfig()

	// 2. 初始化适配器和处理器
	chatHandler := initHandlers(apiKey)

	// 3. 设置路由
	r := router.SetupRouter(chatHandler)

	// 4. 启动服务器
	startServer(r)
}

// initConfig 初始化配置
// 返回API密钥
func initConfig() string {
	apiKey := os.Getenv("GLM_API_KEY")
	if apiKey == "" {
		// 如果环境变量未设置，使用测试密钥
		apiKey = TestAPIKey
		log.Println("⚠️  使用测试API密钥（开发模式）")
	} else {
		log.Println("✓ 使用环境变量API密钥（生产模式）")
	}
	return apiKey
}

// initHandlers 初始化适配器和处理器
// 参数：
//   - apiKey: API密钥
// 返回：
//   - *handler.ChatHandler: 聊天处理器
func initHandlers(apiKey string) *handler.ChatHandler {
	// 创建GLM适配器
	glmAdapter := glm.NewGLMAdapter(apiKey)
	log.Println("✓ GLM适配器初始化成功")

	// 创建适配器管理器
	manager := adapter.NewAdapterManager()

	// 注册GLM适配器（支持多个模型名）
	// GLM-4系列模型都使用同一个适配器实例
	if err := manager.Register("glm-4", glmAdapter); err != nil {
		log.Fatal("❌ 注册glm-4失败:", err)
	}
	if err := manager.Register("glm-4-flash", glmAdapter); err != nil {
		log.Fatal("❌ 注册glm-4-flash失败:", err)
	}
	if err := manager.Register("glm-4-air", glmAdapter); err != nil {
		log.Fatal("❌ 注册glm-4-air失败:", err)
	}

	log.Println("✓ 适配器管理器初始化成功")
	log.Printf("✓ 已注册模型: %v", manager.ListModels())

	// 创建ChatHandler（使用管理器而非单个适配器）
	chatHandler := handler.NewChatHandler(manager)
	log.Println("✓ ChatHandler初始化成功")

	return chatHandler
}

// startServer 启动HTTP服务器
// 参数：
//   - r: Gin路由器
func startServer(r interface{ Run(addr ...string) error }) {
	log.Println("========================================")
	log.Println("🚀 Prism AI Gateway 启动成功")
	log.Println("📍 监听地址: http://localhost:8080")
	log.Println("📚 接口文档:")
	log.Println("   - GET  /           欢迎页面")
	log.Println("   - GET  /health     健康检查")
	log.Println("   - POST /v1/chat/completions  聊天补全")
	log.Println("========================================")

	if err := r.Run(":8080"); err != nil {
		log.Fatal("❌ 启动失败:", err)
	}
}