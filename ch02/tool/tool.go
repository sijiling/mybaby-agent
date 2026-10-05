package tool

import (
	"context"

	"github.com/openai/openai-go/v3"
)

type AgentTool string

const (
	AgentToolRead  AgentTool = "read"
	AgentToolWrite AgentTool = "write"
)

type Tool interface {
	ToolName() AgentTool //获取工具名
	Info() openai.ChatCompletionToolUnionParam
	Execute(ctx context.Context, argsInJson string) (string, error)
}
