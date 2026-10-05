package ch02

import (
	"baby-agent/ch02/tool"
	"baby-agent/shared"
	"context"
	"errors"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Agent struct {
	systemPropmt string //系统提示词
	model        string //模型
	client       openai.Client
	messages     []openai.ChatCompletionMessageParamUnion
	tools        map[tool.AgentTool]tool.Tool
}

func NewAgent(config shared.ModelConfig, systemPrompt string, tools []tool.Tool) *Agent {
	a := Agent{
		systemPropmt: systemPrompt,
		model:        config.Model,
		client:       openai.NewClient(option.WithBaseURL(config.BaseURL), option.WithAPIKey(config.ApiKey)),
		tools:        make(map[tool.AgentTool]tool.Tool),
		messages:     make([]openai.ChatCompletionMessageParamUnion, 0),
	}

	for _, t := range tools {
		a.tools[t.ToolName()] = t
	}

	a.messages = append(a.messages, openai.SystemMessage(systemPrompt))
	return &a
}

func (a *Agent) execute(ctx context.Context, toolName string, argumentsInJson string) (string, error) {
	t, ok := a.tools[tool.AgentTool(toolName)]
	if !ok {
		return "", errors.New("tool not found")
	}
	return t.Execute(ctx, argumentsInJson)
}

func (a *Agent) Run(ctx context.Context, query string) (string, error) {
	a.messages = append(a.messages, openai.UserMessage(query))

	var result string

	for {
		params := openai.ChatCompletionNewParams{
			Model:    a.model,
			Messages: a.messages,
			Tools:    make([]openai.ChatCompletionToolUnionParam, 0),
		}

		for _, t := range a.tools {
			params.Tools = append(params.Tools, t.Info())
		}

		log.Printf("calling llm modle %s...", a.model)
		resp, err := a.client.Chat.Completions.New(ctx, params)
		if err != nil {
			log.Fatalf("call to llm modle failed: %s", err)
			return "", err
		}

		if len(resp.Choices) == 0 {
			log.Printf("no choices returned, resp: %v", err)
			return "", err
		}

		message := resp.Choices[0].Message

		a.messages = append(a.messages, message.ToParam())

		if len(message.ToolCalls) == 0 {
			result = message.Content
			break
		}

		for _, toolCall := range message.ToolCalls {
			toolResult, err := a.execute(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
			if err != nil {
				toolResult = err.Error()
			}

			log.Printf("tool call %s, arguments %s, error: %v", toolCall.Function.Name, toolCall.Function.Arguments, toolResult)

			a.messages = append(a.messages, openai.ToolMessage(toolResult, toolCall.ID))
		}
	}
	return result, nil
}
