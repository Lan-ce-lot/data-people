package cmd

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lan-ce-lot/data-people/config"
	"github.com/Lan-ce-lot/data-people/crawler"
	"github.com/Lan-ce-lot/data-people/models"
	"github.com/Lan-ce-lot/data-people/storage"
	"github.com/Lan-ce-lot/data-people/utils"
	"github.com/spf13/cobra"
)

var (
	startDate string
	endDate   string
	workers   int
)

// crawlCmd represents the crawl command
var crawlCmd = &cobra.Command{
	Use:   "crawl",
	Short: "开始抓取人民日报数据",
	Long: `开始抓取人民日报数据

可以通过以下方式指定抓取范围：
1. 使用配置文件中的设置
2. 通过命令行参数指定具体日期范围

示例：
  data-people crawl --config config.yaml
  data-people crawl --start-date 2025-01-01 --end-date 2025-01-31
  data-people crawl --workers 10`,
	Run: func(cmd *cobra.Command, args []string) {
		runCrawler(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(crawlCmd)

	// Here you will define your flags and configuration settings.
	crawlCmd.Flags().StringVar(&startDate, "start-date", "", "开始日期 (YYYY-MM-DD)")
	crawlCmd.Flags().StringVar(&endDate, "end-date", "", "结束日期 (YYYY-MM-DD)")
	crawlCmd.Flags().IntVar(&workers, "workers", 0, "并发worker数量 (0表示使用配置文件设置)")
}

func runCrawler(_ *cobra.Command, _ []string) {
	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 命令行参数覆盖配置文件设置
	if startDate != "" {
		cfg.DateRange.StartDate = startDate
	}
	if endDate != "" {
		cfg.DateRange.EndDate = endDate
	}
	if workers > 0 {
		cfg.Crawler.Workers = workers
	}

	fmt.Printf("=== %s v%s ===\n", cfg.App.Name, cfg.App.Version)
	fmt.Printf("配置文件: %s\n", configFile)
	fmt.Printf("并发worker数: %d\n", cfg.Crawler.Workers)
	fmt.Printf("请求间隔: %v\n", cfg.Crawler.RequestInterval)
	fmt.Printf("存储类型: %v\n", cfg.Storage.Types)
	if cfg.DateRange.StartDate != "" && cfg.DateRange.EndDate != "" {
		fmt.Printf("日期范围: %s 到 %s\n", cfg.DateRange.StartDate, cfg.DateRange.EndDate)
	} else {
		fmt.Printf("年份范围: %d 到 %d\n", cfg.DateRange.StartYear, cfg.DateRange.EndYear)
	}
	fmt.Println()

	// 创建存储实例
	storages, err := createStorages(cfg)
	if err != nil {
		log.Fatalf("创建存储实例失败: %v", err)
	}
	defer closeStorages(storages)

	// 初始化存储
	for _, store := range storages {
		if err := store.Init(); err != nil {
			log.Fatalf("初始化%s存储失败: %v", store.GetStorageType(), err)
		}
		fmt.Printf("✓ %s存储初始化成功\n", store.GetStorageType())
	}

	// 创建HTTP客户端
	httpClient := crawler.NewHTTPClient(cfg.Crawler.Timeout, cfg.Crawler.UserAgent, cfg.Crawler.BaseCookies)

	// 创建数据解析器
	parser := crawler.NewParser(httpClient)

	// 创建URL构建器
	urlBuilder := utils.NewURLBuilder(cfg.Crawler.BaseSearchURL)

	// 生成日期范围任务
	var dateRanges []utils.DateRange

	// 检查是否设置了具体日期范围
	if cfg.DateRange.StartDate != "" && cfg.DateRange.EndDate != "" {
		var parseErr error
		dateRanges, parseErr = urlBuilder.ParseSpecificDateRange(cfg.DateRange.StartDate, cfg.DateRange.EndDate)
		if parseErr != nil {
			log.Fatalf("解析具体日期范围失败: %v", parseErr)
		}
		fmt.Printf("生成了 %d 个任务 (%s 到 %s)\n",
			len(dateRanges), cfg.DateRange.StartDate, cfg.DateRange.EndDate)
	} else {
		dateRanges = urlBuilder.ParseDateRange(cfg.DateRange.StartYear, cfg.DateRange.EndYear)
		fmt.Printf("生成了 %d 个月份任务 (%d年-%d年)\n",
			len(dateRanges), cfg.DateRange.StartYear, cfg.DateRange.EndYear)
	}

	// 创建统计信息
	stats := &models.CrawlerStats{
		TotalTasks: len(dateRanges),
		StartTime:  time.Now(),
	}

	// 设置信号处理
	signalChan := make(chan os.Signal, 1)
	doneChan := make(chan bool, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动爬虫
	fmt.Println("开始抓取数据...")
	go runCrawlerWorker(cfg, httpClient, parser, urlBuilder, storages, dateRanges, stats, doneChan)

	// 等待完成或中断信号
	select {
	case <-doneChan:
		fmt.Println("\n✓ 抓取任务完成")
	case <-signalChan:
		fmt.Println("\n收到中断信号，正在优雅关闭...")
	}

	// 显示最终统计
	showFinalStats(stats)
}

// runCrawlerWorker 运行爬虫工作程序
func runCrawlerWorker(cfg *config.Config, httpClient *crawler.HTTPClient, parser *crawler.Parser,
	urlBuilder *utils.URLBuilder, storages []storage.Storage,
	dateRanges []utils.DateRange, stats *models.CrawlerStats, doneChan chan bool) {

	defer func() {
		doneChan <- true
	}()

	for i, dateRange := range dateRanges {
		fmt.Printf("[%d/%d] 处理时间段: %s\n", i+1, len(dateRanges), dateRange.String())

		if err := crawlDateRange(cfg, httpClient, parser, urlBuilder, storages, dateRange, stats); err != nil {
			log.Printf("处理时间段失败 [%s]: %v", dateRange.String(), err)
			stats.FailedTasks++
		} else {
			log.Printf("✓ 处理时间段成功 [%s]\n", dateRange.String())
			stats.CompletedTasks++
		}

		// 请求间隔（带随机延迟）
		sleepWithRandomDelay(cfg.Crawler.RequestInterval)
	}

	stats.Duration = time.Since(stats.StartTime)
	if stats.Duration > 0 {
		stats.ArticlesPerSec = float64(stats.TotalArticles) / stats.Duration.Seconds()
	}
}

// crawlDateRange 抓取指定日期范围的数据
func crawlDateRange(cfg *config.Config, httpClient *crawler.HTTPClient, parser *crawler.Parser,
	urlBuilder *utils.URLBuilder, storages []storage.Storage,
	dateRange utils.DateRange, stats *models.CrawlerStats) error {

	pageNo := 1
	hasMore := true

	for hasMore {
		fmt.Printf("  处理第 %d 页\n", pageNo)
		pageHasResults := false

		// 内循环：遍历当前页的所有position (0-19)
		for position := 0; position < 20; position++ {
			// 构建搜索URL
			searchURL, err := urlBuilder.BuildSearchURL(dateRange.Start, dateRange.End, pageNo, position)
			if err != nil {
				return fmt.Errorf("构建搜索URL失败: %v", err)
			}

			fmt.Printf("    请求URL (position=%d): %s\n", position, searchURL)

			// 发送请求，传递页码信息给Cookie
			responseBody, err := httpClient.GetWithRetryAndPageInfo(searchURL, cfg.Crawler.MaxRetries, cfg.Crawler.RequestInterval, pageNo, 20)
			if err != nil {
				return fmt.Errorf("获取搜索结果失败: %v", err)
			}

			fmt.Printf("    响应长度: %d 字节\n", len(responseBody))

			// 解析响应
			response, err := parser.ParseSearchResponse(responseBody, searchURL)
			if err != nil {
				return fmt.Errorf("解析搜索响应失败: %v", err)
			}

			// 检查是否有结果
			if len(response.Data.Results) == 0 {
				fmt.Printf("    position %d 无结果\n", position)
				break // 当前页没有更多结果，跳出内循环
			}

			pageHasResults = true

			// 转换为指针切片
			var articles []*models.Article
			for i := range response.Data.Results {
				articles = append(articles, &response.Data.Results[i])
			}

			log.Printf("    获取到 %d 篇文章 (position=%d)\n", len(articles), position)

			// 保存到各个存储
			for _, store := range storages {
				if err := store.SaveBatch(articles); err != nil {
					log.Printf("保存到%s失败: %v", store.GetStorageType(), err)
				} else {
					fmt.Printf("    ✓ 保存 %d 篇文章到%s (position=%d)\n", len(articles), store.GetStorageType(), position)
				}
			}

			// 更新统计
			stats.TotalArticles += len(articles)

			// position间隔（带随机延迟）
			sleepWithRandomDelay(cfg.Crawler.RequestInterval)
		}

		// 如果当前页没有任何结果，说明没有更多数据了
		if !pageHasResults {
			hasMore = false
		} else {
			// 进入下一页
			pageNo++
			// 页面间隔（带随机延迟）
			sleepWithRandomDelay(cfg.Crawler.RequestInterval)
		}
	}

	return nil
}

// createStorages 创建存储实例
func createStorages(cfg *config.Config) ([]storage.Storage, error) {
	var storages []storage.Storage

	for _, storageType := range cfg.Storage.Types {
		switch storageType {
		case "csv":
			csvStorage := storage.NewCSVStorage(
				cfg.Storage.CSV.OutputDir,
				cfg.Storage.CSV.FilePrefix,
			)
			storages = append(storages, csvStorage)

		case "mysql":
			mysqlStorage := storage.NewMySQLStorage(
				cfg.Storage.MySQL.Host,
				cfg.Storage.MySQL.Port,
				cfg.Storage.MySQL.Username,
				cfg.Storage.MySQL.Password,
				cfg.Storage.MySQL.Database,
				cfg.Storage.MySQL.Charset,
			)
			storages = append(storages, mysqlStorage)

		default:
			return nil, fmt.Errorf("不支持的存储类型: %s", storageType)
		}
	}

	if len(storages) == 0 {
		return nil, fmt.Errorf("没有配置任何存储类型")
	}

	return storages, nil
}

// closeStorages 关闭所有存储
func closeStorages(storages []storage.Storage) {
	for _, store := range storages {
		if err := store.Close(); err != nil {
			log.Printf("关闭%s存储失败: %v", store.GetStorageType(), err)
		}
	}
}

// showFinalStats 显示最终统计信息
func showFinalStats(stats *models.CrawlerStats) {
	fmt.Println("\n=== 抓取统计 ===")
	fmt.Printf("总任务数: %d\n", stats.TotalTasks)
	fmt.Printf("完成任务: %d\n", stats.CompletedTasks)
	fmt.Printf("失败任务: %d\n", stats.FailedTasks)
	fmt.Printf("总文章数: %d\n", stats.TotalArticles)
	fmt.Printf("耗时: %v\n", stats.Duration.Round(time.Second))
	fmt.Printf("平均速度: %.2f 篇/秒\n", stats.ArticlesPerSec)
}

// sleepWithRandomDelay 睡眠指定时间并添加随机延迟
// 随机延迟范围为 [-2ms, +2ms]
func sleepWithRandomDelay(baseDuration time.Duration) {
	// 生成 -2 到 +2 毫秒的随机延迟
	randomMs := rand.Intn(5000) - 2000 // 生成 -2, -1, 0, 1, 2
	randomDelay := time.Duration(randomMs) * time.Millisecond

	totalSleep := baseDuration + randomDelay

	// 确保睡眠时间不会是负数
	if totalSleep < 0 {
		totalSleep = 0
	}

	time.Sleep(totalSleep)
}
