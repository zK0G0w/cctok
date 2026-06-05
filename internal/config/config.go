package config

import (
	"os"
	"path/filepath"
	"strings"
)

// TokenPrice 模型定价，单位：$/1M tokens
type TokenPrice struct {
	Input      float64
	Output     float64
	CacheWrite float64
	CacheRead  float64
}

// ModelPricing 模型前缀与定价的映射
type ModelPricing struct {
	Prefix string
	Price  TokenPrice
}

// Config 运行时配置
type Config struct {
	ClaudeDir  string
	Multiplier float64
	Models     []ModelPricing
}

// Default 返回内置默认配置
func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		ClaudeDir:  filepath.Join(homeDir, ".claude"),
		Multiplier: 1.0,
		Models: []ModelPricing{
			{Prefix: "claude-opus-4", Price: TokenPrice{Input: 15.0, Output: 75.0, CacheWrite: 18.75, CacheRead: 1.50}},
			{Prefix: "claude-sonnet-4", Price: TokenPrice{Input: 3.0, Output: 15.0, CacheWrite: 3.75, CacheRead: 0.30}},
			{Prefix: "claude-haiku-4", Price: TokenPrice{Input: 0.80, Output: 4.0, CacheWrite: 1.0, CacheRead: 0.08}},
			{Prefix: "claude-sonnet-3", Price: TokenPrice{Input: 3.0, Output: 15.0, CacheWrite: 3.75, CacheRead: 0.30}},
			{Prefix: "claude-haiku-3", Price: TokenPrice{Input: 0.25, Output: 1.25, CacheWrite: 0.30, CacheRead: 0.03}},
		},
	}
}

// FindPricing 根据模型名前缀匹配定价，未匹配返回零值
func (c *Config) FindPricing(model string) TokenPrice {
	for _, m := range c.Models {
		if strings.HasPrefix(model, m.Prefix) {
			return m.Price
		}
	}
	return TokenPrice{}
}

// CalculateCost 计算给定 token 用量的费用（美元）
func (c *Config) CalculateCost(model string, input, output, cacheWrite, cacheRead int) float64 {
	price := c.FindPricing(model)
	cost := (float64(input)*price.Input +
		float64(output)*price.Output +
		float64(cacheWrite)*price.CacheWrite +
		float64(cacheRead)*price.CacheRead) / 1_000_000
	return cost * c.Multiplier
}
