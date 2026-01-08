package handler

import (
	"github.com/AtSunset1/prism/internal/adapter"
	"github.com/AtSunset1/prism/internal/model"
	"github.com/gin-gonic/gin"
)

// ChatHandler 处理聊天相关的HTTP请求
type ChatHandler struct {
    adapter adapter.ModelAdapter  // 模型适配器（GLMAdapter）
}

func NewChatHandler(adapter adapter.ModelAdapter) *ChatHandler{
	return &ChatHandler{
		adapter: adapter,
	}
}

// HandleChatCompletion 处理聊天补全请求
// 路由：POST /v1/chat/completions
func (h *ChatHandler) HandleChatCompletion(c *gin.Context){
	var req model.ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // 返回参数错误
        errResp := model.NewInvalidRequestError("无效的请求格式: " + err.Error(), "")
        c.JSON(errResp.GetHTTPStatus(), errResp)
        return
    }

	
    // 2. 判断是否流式请求
    if req.Stream {
        // 处理流式请求
        h.handleStreamResponse(c, &req)
    } else {
        // 处理非流式请求
        h.handleNormalResponse(c, &req)
    }
}

func (h *ChatHandler) handleNormalResponse(c *gin.Context, req *model.ChatRequest){
	resp , err := h.adapter.Chat(c,req)
	if err != nil {
		errResp := model.NewAPIError("模型调用失败: " + err.Error())
		c.JSON(errResp.GetHTTPStatus(), errResp)
        return
	}

	c.JSON(200,resp)
}

// handleStreamResponse 处理流式响应
func (h *ChatHandler) handleStreamResponse(c *gin.Context, req *model.ChatRequest) {
    // 1. 设置SSE响应头
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    // 2. 调用适配器获取流式channel
    streamChan, err := h.adapter.ChatStream(c.Request.Context(), req)
    if err != nil {
        // 流式模式下的错误也要以SSE格式返回
        errResp := model.NewAPIError("模型调用失败: " + err.Error())
        c.SSEvent("error", errResp)
        c.Writer.Flush()
        return
    }

    // 3. 从channel读取数据并发送
    for streamResp := range streamChan {
        // 发送数据chunk
        c.SSEvent("message", streamResp)
        c.Writer.Flush() // 立即发送，不缓存
    }

    // 4. 发送结束标记
    c.SSEvent("message", "[DONE]")
    c.Writer.Flush()
}