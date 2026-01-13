package adapter

import (
	"context"
	"fmt"
	"sync"

	"github.com/AtSunset1/prism/internal/model"
)

// AdapterManager 适配器管理器
// 负责管理多个模型适配器，根据模型名称路由到对应的适配器
//
// 核心功能：
//   1. 注册适配器：将模型名称映射到适配器实例
//   2. 获取适配器：根据模型名称获取对应适配器
//   3. 统一接口：实现ModelAdapter接口，可以透明替换单个适配器
//
// 设计模式：
//   - 注册表模式：维护模型到适配器的映射表
//   - 策略模式：根据model动态选择不同的适配器策略
//   - 适配器模式：Manager本身是一个超级适配器
type AdapterManager struct {
	// adapters 存储模型名称到适配器的映射
	// key: 模型名称（如 "glm-4", "doubao"）
	// value: 适配器实例
	adapters map[string]ModelAdapter

	// mu 读写锁，保护adapters map的并发安全
	// 使用RWMutex而非Mutex：允许多个并发读，提高性能
	mu sync.RWMutex
}

// NewAdapterManager 创建一个新的适配器管理器
// 返回：
//   - *AdapterManager: 管理器实例指针
//
// 示例：
//
//	manager := NewAdapterManager()
//	manager.Register("glm-4", glmAdapter)
func NewAdapterManager() *AdapterManager {
	return &AdapterManager{
		adapters: make(map[string]ModelAdapter),
	}
}

// Register 注册一个适配器
// 参数：
//   - modelName: 模型名称（如 "glm-4", "glm-4-flash"）
//   - adapter: 适配器实例
//
// 返回：
//   - error: 如果参数无效或模型名已存在则返回错误
//
// 示例：
//
//	manager.Register("glm-4", glmAdapter)
//	manager.Register("doubao", doubaoAdapter)
func (m *AdapterManager) Register(modelName string, adapter ModelAdapter) error {
	// 参数验证
	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}
	if adapter == nil {
		return fmt.Errorf("adapter cannot be nil")
	}

	m.mu.Lock()         // 写操作，需要独占锁
	defer m.mu.Unlock()

	// 检查模型是否已注册
	// 注意：这里不允许覆盖已注册的模型，如需覆盖可以去掉此检查
	if _, exists := m.adapters[modelName]; exists {
		return fmt.Errorf("model %s already registered", modelName)
	}

	m.adapters[modelName] = adapter
	return nil
}

// GetAdapter 获取指定模型的适配器
// 参数：
//   - modelName: 模型名称
//
// 返回：
//   - ModelAdapter: 适配器实例
//   - error: 如果模型不存在则返回错误
//
// 示例：
//
//	adapter, err := manager.GetAdapter("glm-4")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (m *AdapterManager) GetAdapter(modelName string) (ModelAdapter, error) {
	m.mu.RLock()        // 读操作，使用读锁，允许并发读
	defer m.mu.RUnlock()

	adapter, exists := m.adapters[modelName]
	if !exists {
		return nil, fmt.Errorf("model %s not found", modelName)
	}

	return adapter, nil
}

// ListModels 列出所有已注册的模型名称
// 返回：
//   - []string: 模型名称列表
//
// 示例：
//
//	models := manager.ListModels()
//	fmt.Printf("可用模型: %v\n", models)
func (m *AdapterManager) ListModels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	models := make([]string, 0, len(m.adapters))
	for modelName := range m.adapters {
		models = append(models, modelName)
	}
	return models
}

// ===== 实现ModelAdapter接口 =====
// AdapterManager本身也实现ModelAdapter接口
// 这样可以作为一个"超级适配器"使用，无缝替换单个适配器

// Chat 非流式聊天接口
// 根据请求中的model字段路由到对应的适配器
//
// 参数：
//   - ctx: 上下文（用于超时控制）
//   - req: 聊天请求
//
// 返回：
//   - *model.ChatResponse: 聊天响应
//   - error: 错误信息
func (m *AdapterManager) Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	// 1. 根据请求中的model字段获取对应的适配器
	adapter, err := m.GetAdapter(req.Model)
	if err != nil {
		return nil, fmt.Errorf("获取适配器失败: %w", err)
	}

	// 2. 调用对应适配器的Chat方法
	return adapter.Chat(ctx, req)
}

// ChatStream 流式聊天接口
// 根据请求中的model字段路由到对应的适配器
//
// 参数：
//   - ctx: 上下文（用于取消流）
//   - req: 聊天请求
//
// 返回：
//   - <-chan *model.StreamResponse: 流式响应channel（只读）
//   - error: 错误信息
func (m *AdapterManager) ChatStream(ctx context.Context, req *model.ChatRequest) (<-chan *model.StreamResponse, error) {
	// 1. 根据请求中的model字段获取对应的适配器
	adapter, err := m.GetAdapter(req.Model)
	if err != nil {
		return nil, fmt.Errorf("获取适配器失败: %w", err)
	}

	// 2. 调用对应适配器的ChatStream方法
	return adapter.ChatStream(ctx, req)
}

// Name 返回管理器名称
// 实现ModelAdapter接口
func (m *AdapterManager) Name() string {
	return "adapter-manager"
}

// HealthCheck 健康检查
// 检查所有已注册适配器的健康状态
//
// 参数：
//   - ctx: 上下文（用于超时控制）
//
// 返回：
//   - error: 如果任何一个适配器不健康则返回错误
func (m *AdapterManager) HealthCheck(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 如果没有注册任何适配器，返回错误
	if len(m.adapters) == 0 {
		return fmt.Errorf("no adapters registered")
	}

	// 检查所有适配器的健康状态
	// 注意：这里是串行检查，如果需要提高性能可以改为并发检查
	for modelName, adapter := range m.adapters {
		if err := adapter.HealthCheck(ctx); err != nil {
			return fmt.Errorf("adapter %s health check failed: %w", modelName, err)
		}
	}

	return nil
}
