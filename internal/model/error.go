package model

import (
	"fmt"
	"net/http"
)

// ErrorResponse 错误响应（OpenAI兼容格式）
// 当请求失败时返回给用户
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	// Message 错误消息（人类可读）
	// 示例："Invalid API key", "Model not found"
	Message string `json:"message"`

	// Type 错误类型
	// 常见类型：
	//   - "invalid_request_error"：请求格式错误
	//   - "authentication_error"：认证失败
	//   - "permission_error"：权限不足
	//   - "not_found_error"：资源不存在
	//   - "rate_limit_error"：速率限制
	//   - "api_error"：API内部错误
	//   - "timeout_error"：超时
	Type string `json:"type"`

	// Param 导致错误的参数名（可选）
	// 示例："model", "messages"
	Param string `json:"param,omitempty"`

	// Code 错误代码（可选）
	// 示例："model_not_found", "insufficient_quota"
	Code string `json:"code,omitempty"`
}

// ===== 预定义错误类型 =====

const (
	ErrorTypeInvalidRequest  = "invalid_request_error"
	ErrorTypeAuthentication  = "authentication_error"
	ErrorTypePermission      = "permission_error"
	ErrorTypeNotFound        = "not_found_error"
	ErrorTypeRateLimit       = "rate_limit_error"
	ErrorTypeAPIError        = "api_error"
	ErrorTypeTimeout         = "timeout_error"
	ErrorTypeServerError     = "server_error"
)

// ===== 构造函数 =====

// NewErrorResponse 创建错误响应
func NewErrorResponse(errType, message string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Type:    errType,
			Message: message,
		},
	}
}

// NewInvalidRequestError 创建无效请求错误
func NewInvalidRequestError(message string, param string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Type:    ErrorTypeInvalidRequest,
			Message: message,
			Param:   param,
		},
	}
}

// NewAuthenticationError 创建认证错误
func NewAuthenticationError(message string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Type:    ErrorTypeAuthentication,
			Message: message,
		},
	}
}

// NewNotFoundError 创建资源不存在错误
func NewNotFoundError(resource string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Type:    ErrorTypeNotFound,
			Message: fmt.Sprintf("%s not found", resource),
			Param:   resource,
		},
	}
}

// NewRateLimitError 创建速率限制错误
func NewRateLimitError(message string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Type:    ErrorTypeRateLimit,
			Message: message,
			Code:    "rate_limit_exceeded",
		},
	}
}

// NewAPIError 创建API错误
func NewAPIError(message string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Type:    ErrorTypeAPIError,
			Message: message,
		},
	}
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(operation string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Type:    ErrorTypeTimeout,
			Message: fmt.Sprintf("%s timeout", operation),
		},
	}
}

// NewServerError 创建服务器错误
func NewServerError(message string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Type:    ErrorTypeServerError,
			Message: message,
		},
	}
}

// ===== 辅助方法 =====

// WithParam 添加参数名
func (e *ErrorResponse) WithParam(param string) *ErrorResponse {
	e.Error.Param = param
	return e
}

// WithCode 添加错误代码
func (e *ErrorResponse) WithCode(code string) *ErrorResponse {
	e.Error.Code = code
	return e
}

// GetHTTPStatus 根据错误类型返回对应的HTTP状态码
func (e *ErrorResponse) GetHTTPStatus() int {
	switch e.Error.Type {
	case ErrorTypeInvalidRequest:
		return http.StatusBadRequest // 400
	case ErrorTypeAuthentication:
		return http.StatusUnauthorized // 401
	case ErrorTypePermission:
		return http.StatusForbidden // 403
	case ErrorTypeNotFound:
		return http.StatusNotFound // 404
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests // 429
	case ErrorTypeTimeout:
		return http.StatusRequestTimeout // 408
	case ErrorTypeServerError, ErrorTypeAPIError:
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError // 500
	}
}

// ===== 常用错误消息 =====

var (
	// ErrMissingModel 缺少model字段
	ErrMissingModel = NewInvalidRequestError("model is required", "model")

	// ErrMissingMessages 缺少messages字段
	ErrMissingMessages = NewInvalidRequestError("messages is required", "messages")

	// ErrEmptyMessages messages为空
	ErrEmptyMessages = NewInvalidRequestError("messages cannot be empty", "messages")

	// ErrInvalidAPIKey 无效的API密钥
	ErrInvalidAPIKey = NewAuthenticationError("Invalid API key")

	// ErrMissingAPIKey 缺少API密钥
	ErrMissingAPIKey = NewAuthenticationError("Missing API key in Authorization header")

	// ErrModelNotFound 模型不存在
	ErrModelNotFound = NewNotFoundError("model")

	// ErrRateLimitExceeded 超过速率限制
	ErrRateLimitExceeded = NewRateLimitError("Rate limit exceeded, please try again later")

	// ErrInternalServer 内部服务器错误
	ErrInternalServer = NewServerError("Internal server error")

	// ErrUpstreamTimeout 上游超时
	ErrUpstreamTimeout = NewTimeoutError("Upstream API call")
)

/*
错误响应示例：

{
  "error": {
    "message": "Invalid API key provided",
    "type": "authentication_error",
    "param": null,
    "code": "invalid_api_key"
  }
}

HTTP状态码对照：
- 400 Bad Request：invalid_request_error
- 401 Unauthorized：authentication_error
- 403 Forbidden：permission_error
- 404 Not Found：not_found_error
- 408 Request Timeout：timeout_error
- 429 Too Many Requests：rate_limit_error
- 500 Internal Server Error：api_error, server_error
*/
