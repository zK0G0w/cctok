package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"cctok/internal/config"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "生成默认配置文件",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.ConfigPath()

		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("配置文件已存在: %s", path)
		}

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}

		if err := os.WriteFile(path, []byte(config.GenerateDefault()), 0644); err != nil {
			return fmt.Errorf("写入配置文件失败: %w", err)
		}

		fmt.Printf("配置文件已生成: %s\n", path)
		fmt.Println("你可以编辑该文件来自定义模型定价和倍率。")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
