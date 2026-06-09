package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"cctok/internal/config"

	"github.com/charmbracelet/huh"
)

func runInteractiveConfig(path string) error {
	cfg := config.Load()

	multiplierStr := fmt.Sprintf("%.1f", cfg.Multiplier)

	var selectedModel string
	modelChoices := make([]huh.Option[string], 0, len(cfg.Models)+1)
	for _, m := range cfg.Models {
		label := fmt.Sprintf("%s (in:$%.2f out:$%.2f)", m.Prefix, m.Price.Input, m.Price.Output)
		modelChoices = append(modelChoices, huh.NewOption(label, m.Prefix))
	}
	modelChoices = append(modelChoices, huh.NewOption("+ 添加新模型", "__new__"))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("全局倍率 (multiplier)").
				Description("默认 1.0，费用 = token × 单价 × 倍率").
				Value(&multiplierStr),
			huh.NewSelect[string]().
				Title("选择要编辑的模型定价").
				Options(modelChoices...).
				Value(&selectedModel),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	newMultiplier, err := strconv.ParseFloat(multiplierStr, 64)
	if err != nil || newMultiplier <= 0 {
		return fmt.Errorf("无效的倍率值: %s", multiplierStr)
	}
	cfg.Multiplier = newMultiplier

	if selectedModel == "__new__" {
		if err := addNewModel(cfg); err != nil {
			return err
		}
	} else if selectedModel != "" {
		if err := editModel(cfg, selectedModel); err != nil {
			return err
		}
	}

	return saveConfig(cfg, path)
}

func editModel(cfg *config.Config, prefix string) error {
	var price config.TokenPrice
	for _, m := range cfg.Models {
		if m.Prefix == prefix {
			price = m.Price
			break
		}
	}

	inputStr := fmt.Sprintf("%.2f", price.Input)
	outputStr := fmt.Sprintf("%.2f", price.Output)
	cacheWriteStr := fmt.Sprintf("%.2f", price.CacheWrite)
	cacheReadStr := fmt.Sprintf("%.4f", price.CacheRead)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(fmt.Sprintf("编辑模型: %s", prefix)).Description("单位: $/1M tokens"),
			huh.NewInput().Title("Input 价格").Value(&inputStr),
			huh.NewInput().Title("Output 价格").Value(&outputStr),
			huh.NewInput().Title("Cache Write 价格").Value(&cacheWriteStr),
			huh.NewInput().Title("Cache Read 价格").Value(&cacheReadStr),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	newPrice, err := parsePrice(inputStr, outputStr, cacheWriteStr, cacheReadStr)
	if err != nil {
		return err
	}

	for i, m := range cfg.Models {
		if m.Prefix == prefix {
			cfg.Models[i].Price = newPrice
			break
		}
	}
	return nil
}

func addNewModel(cfg *config.Config) error {
	var prefix, inputStr, outputStr, cacheWriteStr, cacheReadStr string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title("添加新模型").Description("单位: $/1M tokens"),
			huh.NewInput().Title("模型前缀 (如 claude-opus-4)").Value(&prefix),
			huh.NewInput().Title("Input 价格").Value(&inputStr),
			huh.NewInput().Title("Output 价格").Value(&outputStr),
			huh.NewInput().Title("Cache Write 价格").Value(&cacheWriteStr),
			huh.NewInput().Title("Cache Read 价格").Value(&cacheReadStr),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if prefix == "" {
		return fmt.Errorf("模型前缀不能为空")
	}

	newPrice, err := parsePrice(inputStr, outputStr, cacheWriteStr, cacheReadStr)
	if err != nil {
		return err
	}

	cfg.Models = append(cfg.Models, config.ModelPricing{Prefix: prefix, Price: newPrice})
	return nil
}

func parsePrice(inputStr, outputStr, cacheWriteStr, cacheReadStr string) (config.TokenPrice, error) {
	input, err := strconv.ParseFloat(inputStr, 64)
	if err != nil {
		return config.TokenPrice{}, fmt.Errorf("无效的 Input 价格 %q: %w", inputStr, err)
	}
	output, err := strconv.ParseFloat(outputStr, 64)
	if err != nil {
		return config.TokenPrice{}, fmt.Errorf("无效的 Output 价格 %q: %w", outputStr, err)
	}
	cacheWrite, err := strconv.ParseFloat(cacheWriteStr, 64)
	if err != nil {
		return config.TokenPrice{}, fmt.Errorf("无效的 Cache Write 价格 %q: %w", cacheWriteStr, err)
	}
	cacheRead, err := strconv.ParseFloat(cacheReadStr, 64)
	if err != nil {
		return config.TokenPrice{}, fmt.Errorf("无效的 Cache Read 价格 %q: %w", cacheReadStr, err)
	}
	return config.TokenPrice{Input: input, Output: output, CacheWrite: cacheWrite, CacheRead: cacheRead}, nil
}

func saveConfig(cfg *config.Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content := fmt.Sprintf("# cctok 配置文件\ndata_dir = %q\nmultiplier = %.1f\n\n", "~/.claude", cfg.Multiplier)
	for _, m := range cfg.Models {
		content += fmt.Sprintf("[pricing.%s]\ninput = %.2f\noutput = %.2f\ncache_write = %.4f\ncache_read = %.4f\n\n",
			m.Prefix, m.Price.Input, m.Price.Output, m.Price.CacheWrite, m.Price.CacheRead)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}

	fmt.Printf("\n配置已保存: %s\n", path)
	return nil
}
