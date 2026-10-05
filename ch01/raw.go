package ch01

import (
	"baby-agent/shared"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// --------------------------chapter 01 发起模型请求和流式输出解析
type RequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`

	FinishReason string `json:"finish_reason"` //用于表示模型本轮生成停下的原因 常见参数：stop tool_calls length content_filter null

	ReasoningContent *string `json:"reasoning_content"` //用于承载思维链

	Reasoning *string `json:"reasoning"` //和ReasoningContent一样用于兼容
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIChatCompletionResponse struct {
	Choices []struct {
		Message ResponseMessage `json:"message"`
	} `json:"choices"`

	Useage *Usage `json:"usage,omitempty"`
}

type OpenAIChatCompletionRequest struct {
	Model string `json:"model"`

	Messages []RequestMessage `json:"messages"`

	Stream bool `json:"stream"` //用于表示是否开启流式输出
}

type OpenAIChatCompletionStreamChunk struct {
	Choices []struct {
		Delta ResponseMessage `json:"delta"`
	} `json:"choices"`

	Usage *Usage `json:"usage,omitempty"`
}

func NonStreamingRequestRawHTTP(ctx context.Context, modelConf shared.ModelConfig, query string) {
	client := http.Client{}

	requestBody := OpenAIChatCompletionRequest{
		Messages: []RequestMessage{
			{Role: "user", Content: query},
		},
		Model:  modelConf.Model,
		Stream: false,
	}
	//构造请求
	bodyBytes, _ := json.Marshal(requestBody)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/chat/completions", modelConf.BaseURL), bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+modelConf.ApiKey)

	//发起请求
	httpRep, err := client.Do(httpReq)
	if err != nil {
		log.Fatal("failed to send http request: %v", err)
		return
	}

	defer httpRep.Body.Close()

	if httpRep.StatusCode != 200 {
		log.Fatalf("failed to send http request: %v", httpRep.StatusCode)
		return
	}

	//解析响应
	respBodyBytes, err := io.ReadAll(httpRep.Body)
	if err != nil {
		log.Fatalf("failed to read http response: %v", err)
		return
	}

	resp := OpenAIChatCompletionResponse{}

	if err = json.Unmarshal(respBodyBytes, &resp); err != nil {
		log.Fatalf("failed to unmarshal http response: %v", err)
		return
	}

	if len(resp.Choices) == 0 {
		log.Printf("no choices returned, resp: %v", resp)
		return
	}

	log.Printf("resp content: %s", resp.Choices[0].Message.Content)
	log.Printf("token usage: %+v", resp.Useage)
}

func StreamingRequestRawHttp(ctx context.Context, modelConf shared.ModelConfig, query string) {
	client := http.Client{}

	requestBody := OpenAIChatCompletionRequest{
		Messages: []RequestMessage{
			{Role: "user", Content: query},
		},
		Model:  modelConf.Model,
		Stream: true,
	}

	bodyBytes, _ := json.Marshal(requestBody)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/chat/completions", modelConf.BaseURL), bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+modelConf.ApiKey)

	httpResp, err := client.Do(httpReq)
	if err != nil {
		log.Fatalf("failed to send http request: %v", err)
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		log.Fatalf("failed to send http request: %v", httpResp.StatusCode)
		return
	}

	scanner := bufio.NewScanner(httpResp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			v := strings.TrimPrefix(line, "data:")

			if strings.TrimSpace(v) == "[DONE]" {
				break
			}

			chunk := OpenAIChatCompletionStreamChunk{}
			if err := json.Unmarshal([]byte(v), &chunk); err != nil {
				log.Fatalf("failed to unmarshal chunk: %v", err)
				return
			}
			log.Printf("stream chunk: %s", v)
			if chunk.Usage != nil {
				log.Printf("token usage: %+v", chunk.Usage)
			}
		}
	}
	if scanner.Err() != nil {
		log.Fatalf("failed to read http response: %v", err)
		return
	}
}
