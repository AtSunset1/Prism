package model

import "time"

// ChatResponse 聊天响应（OpenAI兼容格式）
// 网关返回给用户的响应，统一格式
type ChatResponse struct {
	// ID 响应的唯一标识符
	// 格式：chatcmpl-{随机字符串}
	// 示例："chatcmpl-8VwXXXXXXXXXXX"
	ID string `json:"id"`

	// Object 对象类型
	// 固定值："chat.completion"（非流式）或 "chat.completion.chunk"（流式）
	Object string `json:"object"`

	// Created 创建时间戳（Unix时间戳，秒）
	// 示例：1704672000
	Created int64 `json:"created"`

	// Model 实际使用的模型名称
	// 示例："glm-4", "doubao"
	Model string `json:"model"`

	// Choices 生成的回复列表
	// 通常只有1个元素（除非请求中N>1）
	Choices []Choice `json:"choices"`

	// Usage Token使用情况
	// 用于计费和统计
	Usage Usage `json:"usage"`

	// SystemFingerprint 系统指纹（可选）
	// 用于标识后端系统版本，OpenAI用于追踪模型版本
	SystemFingerprint string `json:"system_fingerprint,omitempty"`
}

// Choice 单个回复选项
type Choice struct {
	// Index 回复的索引
	// 从0开始，如果N=1则只有index=0
	Index int `json:"index"`

	// Message AI的完整回复消息（非流式）
	// 流式响应中此字段为空，使用Delta代替
	Message *Message `json:"message,omitempty"`

	// FinishReason 结束原因
	// 可选值：
	//   - "stop"：正常结束（AI认为回复完整）
	//   - "length"：达到max_tokens限制
	//   - "content_filter"：内容被过滤（违规）
	//   - "function_call"：调用了函数（高级功能）
	//   - null：流式响应中，未结束时为null
	FinishReason string `json:"finish_reason"`

	// LogProbs 对数概率信息（高级功能，可选）
	// 用于分析AI生成时的置信度
	LogProbs interface{} `json:"logprobs,omitempty"`
}

// Usage Token使用统计
type Usage struct {
	// PromptTokens 输入消耗的token数
	// 包括：用户消息 + 系统提示 + 对话历史
	PromptTokens int `json:"prompt_tokens"`

	// CompletionTokens 输出消耗的token数
	// AI生成的回复消耗的token
	CompletionTokens int `json:"completion_tokens"`

	// TotalTokens 总token数
	// TotalTokens = PromptTokens + CompletionTokens
	TotalTokens int `json:"total_tokens"`
}

// ===== 辅助方法 =====

// NewChatResponse 创建一个新的ChatResponse
func NewChatResponse(model string, content string) *ChatResponse {
	now := time.Now().Unix()

	return &ChatResponse{
		ID:      generateID(),
		Object:  "chat.completion",
		Created: now,
		Model:   model,
		Choices: []Choice{
			{
				Index: 0,
				Message: &Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     0, // 需要适配器填充
			CompletionTokens: 0, // 需要适配器填充
			TotalTokens:      0, // 需要适配器填充
		},
	}
}

// GetFirstChoice 获取第一个回复（最常用）
func (r *ChatResponse) GetFirstChoice() *Choice {
	if len(r.Choices) > 0 {
		return &r.Choices[0]
	}
	return nil
}

// GetContent 获取第一个回复的内容
func (r *ChatResponse) GetContent() string {
	choice := r.GetFirstChoice()
	if choice != nil && choice.Message != nil {
		return choice.Message.Content
	}
	return ""
}

// IsComplete 检查回复是否完整（finish_reason != null）
func (r *ChatResponse) IsComplete() bool {
	choice := r.GetFirstChoice()
	return choice != nil && choice.FinishReason != ""
}

// generateID 生成响应ID
// 格式：chatcmpl-{时间戳}{随机数}
func generateID() string {
	// 使用时间戳 + 随机数生成ID
	// 实际项目中可以用更复杂的ID生成策略
	now := time.Now().UnixNano()
	return "chatcmpl-" + time.Now().Format("20060102150405") + "-" + string(rune(now%1000000))
}
