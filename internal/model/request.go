package model

// ChatRequest 聊天请求（OpenAI兼容格式）
// 用户发送给网关的请求，网关会转发给不同的AI模型
type ChatRequest struct {
	// ===== 必需字段 =====

	// Model 模型名称
	// 示例：
	//   - "glm-4" - 智谱GLM-4模型
	//   - "doubao" - 字节豆包模型
	//   - "gpt-3.5-turbo" - OpenAI模型（如果接入）
	Model string `json:"model" binding:"required"`

	// Messages 对话消息列表
	// 包含整个对话历史，按时间顺序排列
	// 至少需要1条消息
	Messages []Message `json:"messages" binding:"required,min=1"`

	// ===== 可选字段 =====

	// Temperature 温度参数，控制输出的随机性
	// 范围：0.0 - 2.0
	//   - 0.0：输出最确定、最保守（适合事实性问答）
	//   - 1.0：平衡随机性和确定性（默认值）
	//   - 2.0：输出最随机、最有创意（适合创作）
	// 示例：
	//   - 0.2：写技术文档
	//   - 0.7：日常对话
	//   - 1.5：写诗、讲故事
	Temperature *float64 `json:"temperature,omitempty"`

	// MaxTokens 最大生成的token数量
	// 限制AI回复的长度，避免超长输出
	// 注意：不同模型的token计算方式可能不同
	//   - 中文：约1.5-2个汉字 = 1个token
	//   - 英文：约0.75个单词 = 1个token
	// 示例：
	//   - 100：简短回答
	//   - 500：中等长度回答
	//   - 2000：详细回答
	MaxTokens *int `json:"max_tokens,omitempty"`

	// Stream 是否启用流式响应
	// true：像ChatGPT一样一个字一个字显示（SSE）
	// false：等待完整回复后一次性返回
	// 默认：false
	Stream bool `json:"stream,omitempty"`

	// TopP 核采样参数
	// 范围：0.0 - 1.0
	// 与Temperature二选一使用（不建议同时设置）
	//   - 0.1：只考虑概率最高的10%的词
	//   - 0.5：考虑概率最高的50%的词
	//   - 1.0：考虑所有词（默认）
	TopP *float64 `json:"top_p,omitempty"`

	// N 生成几个候选回复
	// 默认：1
	// 注意：N>1会增加token消耗和响应时间
	// 大部分场景用默认值1即可
	N *int `json:"n,omitempty"`

	// Stop 停止词列表
	// 当AI生成这些词时，立即停止生成
	// 示例：["\n", "用户：", "END"]
	Stop []string `json:"stop,omitempty"`

	// PresencePenalty 存在惩罚
	// 范围：-2.0 - 2.0
	// 正值：减少重复话题，鼓励讨论新话题
	// 负值：允许重复话题
	// 0：不惩罚（默认）
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty 频率惩罚
	// 范围：-2.0 - 2.0
	// 正值：减少重复用词
	// 负值：允许重复用词
	// 0：不惩罚（默认）
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	// User 用户标识符
	// 用于追踪和分析，可选
	// 示例："user-12345"
	User string `json:"user,omitempty"`
}

// Message 单条消息
type Message struct {
	// Role 消息角色
	// 可选值：
	//   - "system"：系统提示，设定AI的行为规则
	//   - "user"：用户输入
	//   - "assistant"：AI回复（用于多轮对话）
	Role string `json:"role" binding:"required,oneof=system user assistant"`

	// Content 消息内容
	// 实际的文本内容
	Content string `json:"content" binding:"required"`

	// Name 消息发送者的名称（可选）
	// 用于多用户场景，区分不同的用户
	// 示例："张三"、"user1"
	Name string `json:"name,omitempty"`
}

// ===== 辅助方法 =====

// GetTemperature 获取Temperature值，如果未设置则返回默认值1.0
func (r *ChatRequest) GetTemperature() float64 {
	if r.Temperature == nil {
		return 1.0
	}
	return *r.Temperature
}

// GetMaxTokens 获取MaxTokens值，如果未设置则返回0（表示无限制）
func (r *ChatRequest) GetMaxTokens() int {
	if r.MaxTokens == nil {
		return 0
	}
	return *r.MaxTokens
}

// GetTopP 获取TopP值，如果未设置则返回默认值1.0
func (r *ChatRequest) GetTopP() float64 {
	if r.TopP == nil {
		return 1.0
	}
	return *r.TopP
}

// GetN 获取N值，如果未设置则返回默认值1
func (r *ChatRequest) GetN() int {
	if r.N == nil {
		return 1
	}
	return *r.N
}

// HasSystemMessage 检查是否包含system消息
func (r *ChatRequest) HasSystemMessage() bool {
	for _, msg := range r.Messages {
		if msg.Role == "system" {
			return true
		}
	}
	return false
}

// GetLastUserMessage 获取最后一条用户消息
func (r *ChatRequest) GetLastUserMessage() *Message {
	for i := len(r.Messages) - 1; i >= 0; i-- {
		if r.Messages[i].Role == "user" {
			return &r.Messages[i]
		}
	}
	return nil
}
