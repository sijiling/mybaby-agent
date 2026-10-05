package main

import (
	"baby-agent/ch01"
	"baby-agent/shared"
	"context"
	"flag"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	useRaw := flag.Bool("raw", false, "use raw http implementation")
	useStream := flag.Bool("stream", false, "use streaming response")
	query := flag.String("q", "hello", "prompt text")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	modelConf := shared.NewModelConfig()

	switch {
	case *useRaw && *useStream:
		ch01.StreamingRequestRawHttp(ctx, modelConf, *query)
	case *useRaw:
		ch01.NonStreamingRequestRawHTTP(ctx, modelConf, *query)
	case *useStream:
		ch01.StreamingRequestRawHttp(ctx, modelConf, *query)
	default:
		ch01.NonStreamingRequestRawHTTP(ctx, modelConf, *query)
	}

}
