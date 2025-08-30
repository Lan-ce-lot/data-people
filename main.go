package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lan-ce-lot/data-people/config"
	"github.com/Lan-ce-lot/data-people/crawler"
	"github.com/Lan-ce-lot/data-people/models"
	"github.com/Lan-ce-lot/data-people/storage"
	"github.com/Lan-ce-lot/data-people/utils"
)

var (
	configPath = flag.String("config", "config.yaml", "配置文件路径")
	showHelp   = flag.Bool("help", false, "显示帮助信息")
)

func main() {
	flag.Parse()

	if *showHelp {
		showUsage()
		return
	}

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	fmt.Printf("=== %s v%s ===\n", cfg.App.Name, cfg.App.Version)
	fmt.Printf("配置文件: %s\n", *configPath)
	fmt.Printf("并发worker数: %d\n", cfg.Crawler.Workers)
	fmt.Printf("请求间隔: %v\n", cfg.Crawler.RequestInterval)
	fmt.Printf("存储类型: %v\n", cfg.Storage.Types)
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
	urlBuilder := utils.NewURLBuilder()

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
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 创建完成通道
	doneChan := make(chan bool, 1)

	// 启动爬虫
	fmt.Println("开始抓取数据...")
	go runCrawler(cfg, httpClient, parser, urlBuilder, storages, dateRanges, stats, doneChan)

	// 等待中断信号或爬虫完成
	select {
	case <-signalChan:
		fmt.Println("\n收到中断信号，正在优雅关闭...")
	case <-doneChan:
		fmt.Println("\n爬虫任务完成")
	}

	// 显示最终统计
	showFinalStats(stats)
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

// runCrawler 运行爬虫
func runCrawler(cfg *config.Config, httpClient *crawler.HTTPClient, parser *crawler.Parser,
	urlBuilder *utils.URLBuilder, storages []storage.Storage,
	dateRanges []utils.DateRange, stats *models.CrawlerStats, doneChan chan bool) {

	defer func() {
		// 确保在函数退出时发送完成信号
		select {
		case doneChan <- true:
		default:
		}
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

		// 请求间隔
		time.Sleep(cfg.Crawler.RequestInterval)
	}

	stats.Duration = time.Since(stats.StartTime)
	if stats.Duration > 0 {
		stats.ArticlesPerSec = float64(stats.TotalArticles) / stats.Duration.Seconds()
	}
	fmt.Print("Done! Press Ctrl+C to stop the crawler...\n")
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

			// position间隔
			time.Sleep(cfg.Crawler.RequestInterval)
		}

		// 如果当前页没有任何结果，说明没有更多数据了
		if !pageHasResults {
			hasMore = false
		} else {
			// 进入下一页
			pageNo++
			// 页面间隔
			time.Sleep(cfg.Crawler.RequestInterval)
		}
	}

	return nil
}

// showFinalStats 显示最终统计信息
func showFinalStats(stats *models.CrawlerStats) {
	fmt.Println("\n=== 抓取统计 ===")
	fmt.Printf("总任务数: %d\n", stats.TotalTasks)
	fmt.Printf("完成任务: %d\n", stats.CompletedTasks)
	fmt.Printf("失败任务: %d\n", stats.FailedTasks)
	fmt.Printf("总文章数: %d\n", stats.TotalArticles)

}

// showUsage 显示使用说明
func showUsage() {
	fmt.Printf(`人民日报数据爬虫

使用方法:
  %s [选项]

选项:
  -config string
        配置文件路径 (默认: "config.yaml")
  -help
        显示此帮助信息

示例:
  %s -config my_config.yaml
  %s -help

`, os.Args[0], os.Args[0], os.Args[0])
}
