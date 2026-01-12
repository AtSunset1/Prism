package handler

import (
	"encoding/json"

	"github.com/AtSunset1/prism/internal/adapter"
	"github.com/AtSunset1/prism/internal/model"
	"github.com/gin-gonic/gin"
)

// ChatHandler 处理聊天相关的HTTP请求
// 职责：
//   - 接收并解析HTTP请求
//   - 调用适配器获取AI响应
//   - 返回标准格式的响应
//   - 处理错误情况
type ChatHandler struct {
	adapter adapter.ModelAdapter // 模型适配器（依赖注入）
}

// NewChatHandler 创建一个新的ChatHandler
// 参数：
//   - adapter: 模型适配器（实现了ModelAdapter接口）
// 返回：
//   - *ChatHandler: ChatHandler实例指针
//
// 示例：
//
//	glmAdapter := glm.NewGLMAdapter(apiKey)
//	handler := NewChatHandler(glmAdapter)
func NewChatHandler(adapter adapter.ModelAdapter) *ChatHandler {
	return &ChatHandler{
		adapter: adapter,
	}
}

// HandleChatCompletion 处理聊天补全请求
// 路由：POST /v1/chat/completions
// 支持流式和非流式两种模式
//
// 请求示例：
//
//	{
//	  "model": "glm-4-flash",
//	  "messages": [{"role": "user", "content": "你好"}],
//	  "stream": false
//	}
func (h *ChatHandler) HandleChatCompletion(c *gin.Context) {
	// 1. 解析请求body为ChatRequest
	var req model.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 请求格式错误（JSON格式不正确或必填字段缺失）
		errResp := model.NewInvalidRequestError("无效的请求格式: "+err.Error(), "body")
		c.JSON(errResp.GetHTTPStatus(), errResp)
		return
	}

	// 2. 判断是否为流式请求
	if req.Stream {
		// 处理流式请求（SSE）
		h.handleStreamResponse(c, &req)
	} else {
		// 处理非流式请求（JSON）
		h.handleNormalResponse(c, &req)
	}
}

// handleNormalResponse 处理非流式响应
// 一次性返回完整的AI回复
func (h *ChatHandler) handleNormalResponse(c *gin.Context, req *model.ChatRequest) {
	// 1. 调用适配器获取响应
	// ⚠️ 重点：传递 c.Request.Context() 而不是 c
	// Context包含超时、取消等控制信息
	resp, err := h.adapter.Chat(c.Request.Context(), req)
	if err != nil {
		// 适配器调用失败（可能是API错误、网络错误、超时等）
		errResp := model.NewAPIError("模型调用失败: " + err.Error())
		c.JSON(errResp.GetHTTPStatus(), errResp)
		return
	}

	// 2. 返回成功响应
	c.JSON(200, resp)
}

// handleStreamResponse 处理流式响应
// 使用SSE（Server-Sent Events）协议逐步返回AI回复
func (h *ChatHandler) handleStreamResponse(c *gin.Context, req *model.ChatRequest) {
	// 1. 设置SSE响应头
	c.Header("Content-Type", "text/event-stream") // 声明SSE格式
	c.Header("Cache-Control", "no-cache")         // 禁止缓存
	c.Header("Connection", "keep-alive")          // 保持连接
	c.Header("X-Accel-Buffering", "no")           // 禁用nginx缓冲
	c.Header("Transfer-Encoding", "chunked")      // 分块传输

	// 2. 调用适配器获取流式channel
	streamChan, err := h.adapter.ChatStream(c.Request.Context(), req)
	if err != nil {
		// 流式调用初始化失败
		// 注意：流式模式下也要以SSE格式返回错误
		h.sendSSEError(c, "模型调用失败: "+err.Error())
		return
	}

	// 3. 从channel读取数据并逐步发送
	// 每次从channel收到一个StreamResponse就立即发送给客户端
	for streamResp := range streamChan {
		// 将StreamResponse序列化为JSON
		data, err := json.Marshal(streamResp)
		if err != nil {
			// JSON序列化失败（理论上不应该发生）
			h.sendSSEError(c, "数据序列化失败: "+err.Error())
			continue
		}

		// 发送SSE数据（标准OpenAI格式）
		// SSE格式：data: {json}\n\n
		c.Writer.Write([]byte("data: "))
		c.Writer.Write(data)
		c.Writer.Write([]byte("\n\n"))
		c.Writer.Flush() // ⚠️ 关键：立即发送，不缓存（实现逐字输出）
	}

	// 4. 发送结束标记
	c.Writer.Write([]byte("data: [DONE]\n\n"))
	c.Writer.Flush()
}

// sendSSEError 以SSE格式发送错误
// 用于流式响应中的错误处理
func (h *ChatHandler) sendSSEError(c *gin.Context, message string) {
	errResp := model.NewAPIError(message)
	data, _ := json.Marshal(errResp)
	c.Writer.Write([]byte("data: "))
	c.Writer.Write(data)
	c.Writer.Write([]byte("\n\n"))
	c.Writer.Flush()
}