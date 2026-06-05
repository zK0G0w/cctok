package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"cctok/internal/config"

	"github.com/spf13/cobra"
)

var configInteractive bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "编辑配置文件",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.ConfigPath()

		// 配置文件不存在则自动生成
		if _, err := os.Stat(path); os.IsNotExist(err) {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}
			if err := os.WriteFile(path, []byte(config.GenerateDefault()), 0644); err != nil {
				return fmt.Errorf("生成配置文件失败: %w", err)
			}
			fmt.Printf("已生成默认配置: %s\n", path)
		}

		if configInteractive {
			return runInteractiveConfig(path)
		}

		return openEditor(path)
	},
}

func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	c := exec.Command(editor, path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func init() {
	configCmd.Flags().BoolVarP(&configInteractive, "interactive", "i", false, "交互式配置向导")
	rootCmd.AddCommand(configCmd)
}
