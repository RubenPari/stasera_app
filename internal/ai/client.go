package ai

import (
	"github.com/sashabaranov/go-openai"
)

// NewGatewayClient returns an OpenAI-compatible client pointed at Vercel AI Gateway.
func NewGatewayClient(apiKey string) *openai.Client {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = "https://ai-gateway.vercel.sh/v1"
	return openai.NewClientWithConfig(cfg)
}
