package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// TokenPrice 模型定价，单位：$/1M tokens
type TokenPrice struct {
	Input      float64 `toml:"input"`
	Output     float64 `toml:"output"`
	CacheWrite float64 `toml:"cache_write"`
	CacheRead  float64 `toml:"cache_read"`
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

// tomlConfig 对应 TOML 文件结构
type tomlConfig struct {
	DataDir    string                `toml:"data_dir"`
	Multiplier float64               `toml:"multiplier"`
	Pricing    map[string]TokenPrice `toml:"pricing"`
}

// Default 返回内置默认配置
func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		ClaudeDir:  filepath.Join(homeDir, ".claude"),
		Multiplier: 1.0,
		Models:     defaultModels(),
	}
}

func defaultModels() []ModelPricing {
	return []ModelPricing{
		{Prefix: "claude-opus-4", Price: TokenPrice{Input: 15.0, Output: 75.0, CacheWrite: 18.75, CacheRead: 1.50}},
		{Prefix: "claude-sonnet-4", Price: TokenPrice{Input: 3.0, Output: 15.0, CacheWrite: 3.75, CacheRead: 0.30}},
		{Prefix: "claude-haiku-4", Price: TokenPrice{Input: 0.80, Output: 4.0, CacheWrite: 1.0, CacheRead: 0.08}},
		{Prefix: "claude-sonnet-3", Price: TokenPrice{Input: 3.0, Output: 15.0, CacheWrite: 3.75, CacheRead: 0.30}},
		{Prefix: "claude-haiku-3", Price: TokenPrice{Input: 0.25, Output: 1.25, CacheWrite: 0.30, CacheRead: 0.03}},
	}
}

// FindPricing 根据模型名前缀匹配定价，优先匹配最长前缀（最精确），未匹配返回零值
func (c *Config) FindPricing(model string) TokenPrice {
	bestIdx := -1
	bestLen := 0
	for i, m := range c.Models {
		if strings.HasPrefix(model, m.Prefix) && len(m.Prefix) > bestLen {
			bestIdx = i
			bestLen = len(m.Prefix)
		}
	}
	if bestIdx >= 0 {
		return c.Models[bestIdx].Price
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

// ConfigPath 返回配置文件路径
func ConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".cctok", "config.toml")
}

// Load 尝试加载配置文件，不存在则返回默认配置
func Load() *Config {
	cfg := Default()
	path := ConfigPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg
	}

	var tc tomlConfig
	if _, err := toml.DecodeFile(path, &tc); err != nil {
		fmt.Fprintf(os.Stderr, "warning: 配置文件解析失败，使用默认配置: %v\n", err)
		return cfg
	}

	if tc.DataDir != "" {
		expanded := tc.DataDir
		if strings.HasPrefix(expanded, "~/") {
			homeDir, _ := os.UserHomeDir()
			expanded = filepath.Join(homeDir, expanded[2:])
		}
		cfg.ClaudeDir = expanded
	}
	if tc.Multiplier > 0 {
		cfg.Multiplier = tc.Multiplier
	}
	if len(tc.Pricing) > 0 {
		models := make([]ModelPricing, 0, len(tc.Pricing))
		for prefix, price := range tc.Pricing {
			models = append(models, ModelPricing{Prefix: prefix, Price: price})
		}
		cfg.Models = models
	}

	return cfg
}

// GenerateDefault 生成默认配置文件内容
func GenerateDefault() string {
	return `# cctok 配置文件
# Claude Code 数据目录（默认 ~/.claude）
data_dir = "~/.claude"

# 全局倍率（如 Max plan 5x，默认 1.0 为原始 API 定价）
multiplier = 1.0

# 模型定价（单位：$/1M tokens）
# 前缀匹配：claude-opus-4 会匹配 claude-opus-4-6、claude-opus-4-7 等
[pricing.claude-opus-4]
input = 15.0
output = 75.0
cache_write = 18.75
cache_read = 1.50

[pricing.claude-sonnet-4]
input = 3.0
output = 15.0
cache_write = 3.75
cache_read = 0.30

[pricing.claude-haiku-4]
input = 0.80
output = 4.0
cache_write = 1.0
cache_read = 0.08

[pricing.claude-sonnet-3]
input = 3.0
output = 15.0
cache_write = 3.75
cache_read = 0.30

[pricing.claude-haiku-3]
input = 0.25
output = 1.25
cache_write = 0.30
cache_read = 0.03
`
}
