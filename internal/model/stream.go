package model

import "time"

// StreamResponse 流式响应（OpenAI兼容格式）
// SSE（Server-Sent Events）格式的单个数据块
// 完整的流式响应 = 多个StreamResponse的序列
type StreamResponse struct {
	// ID 响应的唯一标识符（整个流式会话保持不变）
	// 格式：chatcmpl-{随机字符串}
	ID string `json:"id"`

	// Object 对象类型
	// 固定值："chat.completion.chunk"（注意和非流式的区别）
	Object string `json:"object"`

	// Created 创建时间戳（Unix时间戳，秒）
	Created int64 `json:"created"`

	// Model 实际使用的模型名称
	Model string `json:"model"`

	// Choices 生成的回复增量列表
	Choices []StreamChoice `json:"choices"`

	// SystemFingerprint 系统指纹（可选）
	SystemFingerprint string `json:"system_fingerprint,omitempty"`
}

// StreamChoice 流式回复选项
type StreamChoice struct {
	// Index 回复的索引
	Index int `json:"index"`

	// Delta 增量内容（和非流式的Message不同）
	// 每个chunk只包含新增的内容
	Delta StreamDelta `json:"delta"`

	// FinishReason 结束原因
	// 大部分chunk中为null，只有最后一个chunk有值
	//   - null：还在生成中
	//   - "stop"：正常结束
	//   - "length"：达到长度限制
	FinishReason *string `json:"finish_reason"`

	// LogProbs 对数概率信息（可选）
	LogProbs interface{} `json:"logprobs,omitempty"`
}

// StreamDelta 流式增量内容
// 和Message的区别：
//   - Message：完整的消息
//   - Delta：增量的内容片段
type StreamDelta struct {
	// Role 角色（只在第一个chunk中出现）
	// 值："assistant"
	Role string `json:"role,omitempty"`

	// Content 增量内容
	// 示例流程：
	//   chunk1: "你"
	//   chunk2: "好"
	//   chunk3: "，"
	//   chunk4: "我"
	//   chunk5: "是"
	//   chunk6: "AI"
	Content string `json:"content,omitempty"`
}

// ===== 辅助方法 =====

// NewStreamResponse 创建一个新的流式响应
func NewStreamResponse(id string, model string, delta string, isFirst bool) *StreamResponse {
	now := time.Now().Unix()

	choice := StreamChoice{
		Index: 0,
		Delta: StreamDelta{
			Content: delta,
		},
		FinishReason: nil,
	}

	// 第一个chunk需要包含role
	if isFirst {
		choice.Delta.Role = "assistant"
		choice.Delta.Content = "" // 第一个chunk通常content为空
	}

	return &StreamResponse{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: now,
		Model:   model,
		Choices: []StreamChoice{choice},
	}
}

// NewStreamEndResponse 创建流式结束响应
func NewStreamEndResponse(id string, model string, finishReason string) *StreamResponse {
	now := time.Now().Unix()
	reason := finishReason

	return &StreamResponse{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: now,
		Model:   model,
		Choices: []StreamChoice{
			{
				Index:        0,
				Delta:        StreamDelta{}, // 空delta
				FinishReason: &reason,
			},
		},
	}
}

// IsFirst 判断是否是第一个chunk（包含role）
func (s *StreamResponse) IsFirst() bool {
	if len(s.Choices) > 0 {
		return s.Choices[0].Delta.Role != ""
	}
	return false
}

// IsEnd 判断是否是最后一个chunk（有finish_reason）
func (s *StreamResponse) IsEnd() bool {
	if len(s.Choices) > 0 {
		return s.Choices[0].FinishReason != nil
	}
	return false
}

// GetContent 获取增量内容
func (s *StreamResponse) GetContent() string {
	if len(s.Choices) > 0 {
		return s.Choices[0].Delta.Content
	}
	return ""
}

// GetFinishReason 获取结束原因
func (s *StreamResponse) GetFinishReason() string {
	if len(s.Choices) > 0 && s.Choices[0].FinishReason != nil {
		return *s.Choices[0].FinishReason
	}
	return ""
}

// ===== SSE格式化 =====

// ToSSEData 转换为SSE格式的数据
// SSE格式：data: {JSON}\n\n
func (s *StreamResponse) ToSSEData() string {
	// 在实际使用中，需要先序列化为JSON
	// 这里只是示意格式
	return "data: " + "{JSON}" + "\n\n"
}

// SSEDoneMessage SSE流结束标记
const SSEDoneMessage = "data: [DONE]\n\n"

/*
完整的流式响应示例：

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1704672000,"model":"glm-4","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1704672000,"model":"glm-4","choices":[{"index":0,"delta":{"content":"你"},"finish_reason":null}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1704672000,"model":"glm-4","choices":[{"index":0,"delta":{"content":"好"},"finish_reason":null}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1704672000,"model":"glm-4","choices":[{"index":0,"delta":{"content":"！"},"finish_reason":null}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1704672000,"model":"glm-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]

*/
