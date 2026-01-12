package router

import (
	"net/http"

	"github.com/AtSunset1/prism/internal/handler"
	"github.com/gin-gonic/gin"
)

// SetupRouter 配置并返回Gin路由器
// 参数：
//   - chatHandler: 聊天处理器
// 返回：
//   - *gin.Engine: 配置好的Gin路由器
func SetupRouter(chatHandler *handler.ChatHandler) *gin.Engine {
	// 创建Gin路由器（包含Logger和Recovery中间件）
	r := gin.Default()

	// 注册路由
	registerRoutes(r, chatHandler)

	return r
}

// registerRoutes 注册所有路由
func registerRoutes(r *gin.Engine, chatHandler *handler.ChatHandler) {
	// ========== 基础路由 ==========

	// 欢迎页面
	r.GET("/", handleWelcome)

	// 健康检查
	r.GET("/health", handleHealthCheck)

	// ========== OpenAI兼容API路由 ==========

	// v1版本API组
	v1 := r.Group("/v1")
	{
		// 聊天补全接口（核心功能）
		v1.POST("/chat/completions", chatHandler.HandleChatCompletion)
	}
}

// handleWelcome 欢迎页面处理器
func handleWelcome(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to Prism AI Gateway!",
		"version": "0.1.0",
		"status":  "running",
		"description": "OpenAI-compatible AI Gateway for multiple LLM providers",
		"endpoints": gin.H{
			"health": "GET /health",
			"chat":   "POST /v1/chat/completions",
		},
	})
}

// handleHealthCheck 健康检查处理器
func handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
