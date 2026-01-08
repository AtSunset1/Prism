package adapter

import (
	"context"

	"github.com/AtSunset1/prism/internal/model"
)

type ModelAdapter interface {
	//普通调用
	Chat(ctx context.Context, req *model.ChatRequest)(*model.ChatResponse, error)
	//流式调用 SSE
	ChatStream(ctx context.Context, req *model.ChatRequest)(<-chan *model.StreamResponse, error)
	//获取adapter名字
    Name() string
	//健康检查
    HealthCheck(ctx context.Context) error
}