package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	App       AppConfig       `mapstructure:"app" yaml:"app"`
	Crawler   CrawlerConfig   `mapstructure:"crawler" yaml:"crawler"`
	DateRange DateRangeConfig `mapstructure:"date_range" yaml:"date_range"`
	Storage   StorageConfig   `mapstructure:"storage" yaml:"storage"`
	Logging   LoggingConfig   `mapstructure:"logging" yaml:"logging"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	Name    string `mapstructure:"name" yaml:"name"`
	Version string `mapstructure:"version" yaml:"version"`
}

// CrawlerConfig 爬虫配置
type CrawlerConfig struct {
	Workers         int           `mapstructure:"workers" yaml:"workers"`
	RequestInterval time.Duration `mapstructure:"request_interval" yaml:"request_interval"`
	Timeout         time.Duration `mapstructure:"timeout" yaml:"timeout"`
	MaxRetries      int           `mapstructure:"max_retries" yaml:"max_retries"`
	UserAgent       string        `mapstructure:"user_agent" yaml:"user_agent"`
	BaseCookies     string        `mapstructure:"base_cookies" yaml:"base_cookies"`       // 基础Cookie，不包含页码信息
	BaseSearchURL   string        `mapstructure:"base_search_url" yaml:"base_search_url"` // 基础搜索URL
}

// DateRangeConfig 日期范围配置
type DateRangeConfig struct {
	StartYear int    `mapstructure:"start_year" yaml:"start_year"`
	EndYear   int    `mapstructure:"end_year" yaml:"end_year"`
	StartDate string `mapstructure:"start_date" yaml:"start_date"` // 具体开始日期 YYYY-MM-DD
	EndDate   string `mapstructure:"end_date" yaml:"end_date"`     // 具体结束日期 YYYY-MM-DD
}

// StorageConfig 存储配置
type StorageConfig struct {
	Types []string    `mapstructure:"types" yaml:"types"`
	CSV   CSVConfig   `mapstructure:"csv" yaml:"csv"`
	MySQL MySQLConfig `mapstructure:"mysql" yaml:"mysql"`
}

// CSVConfig CSV存储配置
type CSVConfig struct {
	OutputDir  string `mapstructure:"output_dir" yaml:"output_dir"`
	FilePrefix string `mapstructure:"file_prefix" yaml:"file_prefix"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host         string `mapstructure:"host" yaml:"host"`
	Port         int    `mapstructure:"port" yaml:"port"`
	Username     string `mapstructure:"username" yaml:"username"`
	Password     string `mapstructure:"password" yaml:"password"`
	Database     string `mapstructure:"database" yaml:"database"`
	Charset      string `mapstructure:"charset" yaml:"charset"`
	MaxOpenConns int    `mapstructure:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `mapstructure:"level" yaml:"level"`
	File       string `mapstructure:"file" yaml:"file"`
	MaxSize    int    `mapstructure:"max_size" yaml:"max_size"`
	MaxBackups int    `mapstructure:"max_backups" yaml:"max_backups"`
	MaxAge     int    `mapstructure:"max_age" yaml:"max_age"`
}

// LoadConfig 使用Viper加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 设置默认值
	setDefaults()

	// 配置Viper
	if configPath != "" {
		// 使用指定的配置文件
		viper.SetConfigFile(configPath)
	} else {
		// 搜索配置文件
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.data-people")
		viper.AddConfigPath("/etc/data-people/")
	}

	// 启用环境变量支持
	viper.AutomaticEnv()
	viper.SetEnvPrefix("DATA_PEOPLE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，创建默认配置
			fmt.Printf("配置文件不存在，创建默认配置: %s\n", configPath)
			if err := CreateDefaultConfig(configPath); err != nil {
				return nil, fmt.Errorf("创建默认配置文件失败: %v", err)
			}
			// 重新读取创建的配置文件
			if err := viper.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("读取配置文件失败: %v", err)
			}
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %v", err)
		}
	}

	// 将配置解码到结构体
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	fmt.Printf("✓ 配置文件加载成功: %s\n", viper.ConfigFileUsed())
	return &config, nil
}

// GetConfig 获取当前Viper配置实例（用于动态获取配置值）
func GetConfig() *viper.Viper {
	return viper.GetViper()
}

// GetString 获取字符串配置值
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt 获取整数配置值
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetBool 获取布尔配置值
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetDuration 获取时间间隔配置值
func GetDuration(key string) time.Duration {
	return viper.GetDuration(key)
}

// ReloadConfig 重新加载配置（支持热重载）
func ReloadConfig() error {
	return viper.ReadInConfig()
}

// WatchConfig 监控配置文件变化并自动重载
func WatchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("配置文件发生变化: %s\n", e.Name)
	})
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

// setDefaults 使用Viper设置默认值
func setDefaults() {
	// App默认值
	viper.SetDefault("app.name", "人民日报爬虫")
	viper.SetDefault("app.version", "1.0.0")

	// Crawler默认值
	viper.SetDefault("crawler.workers", 5)
	viper.SetDefault("crawler.request_interval", "1s")
	viper.SetDefault("crawler.timeout", "30s")
	viper.SetDefault("crawler.max_retries", 3)
	viper.SetDefault("crawler.user_agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	viper.SetDefault("crawler.base_cookies", "")
	viper.SetDefault("crawler.base_search_url", "http://paper.people.com.cn/rmrb/pc/layout/")

	// DateRange默认值
	viper.SetDefault("date_range.start_year", 1949)
	viper.SetDefault("date_range.end_year", 2025)
	viper.SetDefault("date_range.start_date", "")
	viper.SetDefault("date_range.end_date", "")

	// Storage默认值
	viper.SetDefault("storage.types", []string{"csv", "mysql"})

	// CSV默认值
	viper.SetDefault("storage.csv.output_dir", "./data")
	viper.SetDefault("storage.csv.file_prefix", "articles")

	// MySQL默认值
	viper.SetDefault("storage.mysql.host", "localhost")
	viper.SetDefault("storage.mysql.port", 3306)
	viper.SetDefault("storage.mysql.username", "root")
	viper.SetDefault("storage.mysql.password", "password")
	viper.SetDefault("storage.mysql.database", "people_daily")
	viper.SetDefault("storage.mysql.charset", "utf8mb4")
	viper.SetDefault("storage.mysql.max_open_conns", 10)
	viper.SetDefault("storage.mysql.max_idle_conns", 5)

	// Logging默认值
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "./logs/crawler.log")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)
}
