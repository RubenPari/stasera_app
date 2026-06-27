package ai

import (
	"github.com/sashabaranov/go-openai"
)

// NewGatewayClient returns an OpenAI-compatible client pointed at the given base URL.
// The baseURL should be provided from config (default https://ai-gateway.vercel.sh/v1).
func NewGatewayClient(apiKey, baseURL string) *openai.Client {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL
	return openai.NewClientWithConfig(cfg)
}