package main

import (
	"baby-agent/ch02"
	"baby-agent/ch02/tool"
	"baby-agent/shared"
	"context"
	"flag"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	qurey := flag.String("q", "hello", "prompt text")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	modelConf := shared.NewModelConfig()

	agent := ch02.NewAgent(modelConf, ch02.CodingAgentSystemPrompt, []tool.Tool{
		tool.NewReadTool(),
	})

	result, err := agent.Run(ctx, *qurey)
	if err != nil {
		log.Fatal("agent run error: %v", err)
		return
	}

	log.Printf("result: %s", result)
}
