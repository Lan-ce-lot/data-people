package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	App       AppConfig       `yaml:"app"`
	Crawler   CrawlerConfig   `yaml:"crawler"`
	DateRange DateRangeConfig `yaml:"date_range"`
	Storage   StorageConfig   `yaml:"storage"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// CrawlerConfig 爬虫配置
type CrawlerConfig struct {
	Workers         int           `yaml:"workers"`
	RequestInterval time.Duration `yaml:"request_interval"`
	Timeout         time.Duration `yaml:"timeout"`
	MaxRetries      int           `yaml:"max_retries"`
	UserAgent       string        `yaml:"user_agent"`
	BaseCookies     string        `yaml:"base_cookies"` // 基础Cookie，不包含页码信息
}

// DateRangeConfig 日期范围配置
type DateRangeConfig struct {
	StartYear int    `yaml:"start_year"`
	EndYear   int    `yaml:"end_year"`
	StartDate string `yaml:"start_date"` // 具体开始日期 YYYY-MM-DD
	EndDate   string `yaml:"end_date"`   // 具体结束日期 YYYY-MM-DD
}

// StorageConfig 存储配置
type StorageConfig struct {
	Types []string    `yaml:"types"`
	CSV   CSVConfig   `yaml:"csv"`
	MySQL MySQLConfig `yaml:"mysql"`
}

// CSVConfig CSV存储配置
type CSVConfig struct {
	OutputDir  string `yaml:"output_dir"`
	FilePrefix string `yaml:"file_prefix"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	Charset      string `yaml:"charset"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 如果配置文件不存在，创建默认配置文件
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := CreateDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %v", err)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 设置默认值
	setDefaults(&config)

	return &config, nil
}

// CreateDefaultConfig 创建默认配置文件
func CreateDefaultConfig(configPath string) error {
	config := getDefaultConfig()

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化默认配置失败: %v", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:    "人民日报爬虫",
			Version: "1.0.0",
		},
		Crawler: CrawlerConfig{
			Workers:         5,
			RequestInterval: 1000 * time.Millisecond,
			Timeout:         30 * time.Second,
			MaxRetries:      3,
			UserAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		},
		DateRange: DateRangeConfig{
			StartYear: 1949,
			EndYear:   2025,
		},
		Storage: StorageConfig{
			Types: []string{"csv", "mysql"},
			CSV: CSVConfig{
				OutputDir:  "./data",
				FilePrefix: "articles",
			},
			MySQL: MySQLConfig{
				Host:         "localhost",
				Port:         3306,
				Username:     "root",
				Password:     "password",
				Database:     "people_daily",
				Charset:      "utf8mb4",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			File:       "./logs/crawler.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		},
	}
}

// setDefaults 设置默认值
func setDefaults(config *Config) {
	if config.Crawler.Workers <= 0 {
		config.Crawler.Workers = 5
	}
	if config.Crawler.RequestInterval <= 0 {
		config.Crawler.RequestInterval = 1000 * time.Millisecond
	}
	if config.Crawler.Timeout <= 0 {
		config.Crawler.Timeout = 30 * time.Second
	}
	if config.Crawler.MaxRetries <= 0 {
		config.Crawler.MaxRetries = 3
	}
	if config.Crawler.UserAgent == "" {
		config.Crawler.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
	}
}
