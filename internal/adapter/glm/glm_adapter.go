package glm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AtSunset1/prism/internal/model"
)

// GLM API 默认配置
const (
	// DefaultGLMURL GLM API 默认地址
	DefaultGLMURL = "https://open.bigmodel.cn/api/paas/v4/chat/completions"

	// DefaultTimeout 默认超时时间
	DefaultTimeout = 30 * time.Second

	// GLMName 适配器名称
	GLMName = "glm"
)

// GLMAdapter 智谱GLM适配器
// 实现 ModelAdapter 接口，用于调用智谱GLM API
type GLMAdapter struct {
	// apiKey API密钥
	apiKey string

	// baseURL API基础URL（默认使用 DefaultGLMURL）
	baseURL string

	// client HTTP客户端（复用连接，提高性能）
	client *http.Client

	// timeout 请求超时时间
	timeout time.Duration
}

// NewGLMAdapter 创建GLM适配器实例
// 参数：
//   - apiKey: 智谱API密钥
// 返回：
//   - *GLMAdapter: 适配器实例
func NewGLMAdapter(apiKey string) *GLMAdapter {
	return &GLMAdapter{
		apiKey:  apiKey,
		baseURL: DefaultGLMURL,
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
		timeout: DefaultTimeout,
	}
}

// NewGLMAdapterWithConfig 使用自定义配置创建GLM适配器
// 参数：
//   - apiKey: 智谱API密钥
//   - baseURL: 自定义API地址（如果为空则使用默认值）
//   - timeout: 超时时间（如果为0则使用默认值）
func NewGLMAdapterWithConfig(apiKey, baseURL string, timeout time.Duration) *GLMAdapter {
	if baseURL == "" {
		baseURL = DefaultGLMURL
	}
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	return &GLMAdapter{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Name 返回适配器名称
// 实现 ModelAdapter 接口
func (a *GLMAdapter) Name() string {
	return GLMName
}

// Chat 非流式聊天接口
// 实现 ModelAdapter 接口
// 参数：
//   - ctx: 上下文（用于超时控制）
//   - req: 聊天请求（OpenAI兼容格式）
// 返回：
//   - *model.ChatResponse: 聊天响应
//   - error: 错误信息
func (a *GLMAdapter) Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	// 1. 构造请求体（GLM格式与我们的模型完全兼容）
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 2. 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 3. 设置请求头
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// 4. 发送请求
	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// 5. 读取响应体
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	// 6. 检查HTTP状态码
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GLM API error (status %d): %s", httpResp.StatusCode, string(respBody))
	}

	// 7. 解析响应（GLM格式与我们的模型兼容）
	var chatResp model.ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	return &chatResp, nil
}

// ChatStream 流式聊天接口
// 实现 ModelAdapter 接口
// 参数：
//   - ctx: 上下文（用于取消流）
//   - req: 聊天请求（OpenAI兼容格式）
// 返回：
//   - <-chan *model.StreamResponse: 流式响应channel（只读）
//   - error: 错误信息
func (a *GLMAdapter) ChatStream(ctx context.Context, req *model.ChatRequest) (<-chan *model.StreamResponse, error) {
	// 1. 强制启用流式模式
	req.Stream = true

	// 2. 构造请求体
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 3. 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 4. 设置请求头（流式请求需要 Accept: text/event-stream）
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	// 5. 发送请求
	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	// 6. 检查HTTP状态码
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		return nil, fmt.Errorf("GLM API error (status %d): %s", httpResp.StatusCode, string(body))
	}

	// 7. 创建channel用于传递流式响应
	streamChan := make(chan *model.StreamResponse, 10) // 带缓冲，避免阻塞

	// 8. 启动goroutine处理流式响应
	go func() {
		defer httpResp.Body.Close()
		defer close(streamChan) // 完成后关闭channel

		// 使用 Scanner 逐行读取 SSE 数据
		scanner := bufio.NewScanner(httpResp.Body)

		for scanner.Scan() {
			line := scanner.Text()

			// SSE 格式：data: {json}
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				// 跳过结束标记
				if data == "[DONE]" {
					break
				}

				// 解析JSON为StreamResponse
				var streamResp model.StreamResponse
				if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
					// 解析失败，忽略这条数据
					continue
				}

				// 发送到channel（检查context是否已取消）
				select {
				case streamChan <- &streamResp:
					// 成功发送
				case <-ctx.Done():
					// context已取消，退出
					return
				}
			}
		}

		// 检查扫描错误
		if err := scanner.Err(); err != nil {
			// 流读取失败，但无法通过channel传递错误
			// 可以考虑在最后发送一个特殊的错误响应
			return
		}
	}()

	return streamChan, nil
}

// HealthCheck 健康检查
// 实现 ModelAdapter 接口
// 发送一个简单的测试请求，验证API是否可用
func (a *GLMAdapter) HealthCheck(ctx context.Context) error {
	// 构造简单的测试请求
	testReq := &model.ChatRequest{
		Model: "glm-4",
		Messages: []model.Message{
			{
				Role:    "user",
				Content: "hi",
			},
		},
		MaxTokens: intPtr(5), // 限制token数量，节省配额
	}

	// 调用Chat接口
	_, err := a.Chat(ctx, testReq)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// intPtr 辅助函数：返回int指针
func intPtr(i int) *int {
	return &i
}
