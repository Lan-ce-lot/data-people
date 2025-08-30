package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "data-people",
	Short: "人民日报数据爬虫",
	Long: `人民日报数据爬虫 - 一个用于抓取人民日报历史文章数据的工具

支持功能：
- 抓取指定时间范围的文章数据
- 支持CSV和MySQL双重存储
- 支持并发抓取和错误重试
- 支持断点续传和增量更新

示例用法：
  data-people crawl --config config.yaml
  data-people crawl --start-date 2025-01-01 --end-date 2025-01-31
  data-people version`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "config.yaml", "配置文件路径")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
