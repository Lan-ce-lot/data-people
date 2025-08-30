package cmd

import (
	"fmt"

	"github.com/Lan-ce-lot/data-people/config"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  `显示程序的版本信息`,
	Run: func(cmd *cobra.Command, args []string) {
		showVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func showVersion() {
	// 尝试加载配置文件来获取版本信息，如果失败则使用默认值
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("人民日报数据抓取工具 v1.0.0\n")
		fmt.Printf("Go版本: go1.20+\n")
		return
	}

	fmt.Printf("%s v%s\n", cfg.App.Name, cfg.App.Version)
	fmt.Printf("Go版本: go1.20+\n")
}
