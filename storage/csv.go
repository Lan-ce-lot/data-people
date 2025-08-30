package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/Lan-ce-lot/data-people/models"
)

// CSVStorage CSV存储实现
type CSVStorage struct {
	outputDir  string
	filePrefix string
	mu         sync.Mutex
	files      map[string]*os.File
	writers    map[string]*csv.Writer
}

// NewCSVStorage 创建CSV存储实例
func NewCSVStorage(outputDir, filePrefix string) *CSVStorage {
	return &CSVStorage{
		outputDir:  outputDir,
		filePrefix: filePrefix,
		files:      make(map[string]*os.File),
		writers:    make(map[string]*csv.Writer),
	}
}

// Init 初始化CSV存储
func (c *CSVStorage) Init() error {
	// 创建输出目录
	if err := os.MkdirAll(c.outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}
	return nil
}

// Save 保存单个文章
func (c *CSVStorage) Save(article *models.Article) error {
	return c.SaveBatch([]*models.Article{article})
}

// SaveBatch 批量保存文章
func (c *CSVStorage) SaveBatch(articles []*models.Article) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 按月份分组文章
	monthlyGroups := make(map[string][]*models.Article)
	for _, article := range articles {
		monthKey := article.PublishDate.Format("200601") // YYYYMM
		monthlyGroups[monthKey] = append(monthlyGroups[monthKey], article)
	}

	// 为每个月份写入文件
	for monthKey, monthArticles := range monthlyGroups {
		if err := c.writeToFile(monthKey, monthArticles); err != nil {
			return fmt.Errorf("写入CSV文件失败 [%s]: %v", monthKey, err)
		}
	}

	return nil
}

// writeToFile 写入指定月份的文件
func (c *CSVStorage) writeToFile(monthKey string, articles []*models.Article) error {
	writer, err := c.getWriter(monthKey)
	if err != nil {
		return err
	}

	for _, article := range articles {
		record := []string{
			strconv.Itoa(article.ID),
			article.URL,
			article.Title,
			article.Subtitle,
			c.escapeCSVField(article.Raw),
			article.PublishDate.Format("2006-01-02 15:04:05"),
			article.Edition,
			article.Type,
			c.escapeCSVField(article.Content),
			article.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("写入CSV记录失败: %v", err)
		}
	}

	// 刷新缓冲区
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("刷新CSV缓冲区失败: %v", err)
	}

	return nil
}

// getWriter 获取指定月份的CSV写入器
func (c *CSVStorage) getWriter(monthKey string) (*csv.Writer, error) {
	if writer, exists := c.writers[monthKey]; exists {
		return writer, nil
	}

	// 创建新的文件和写入器
	filename := fmt.Sprintf("%s_%s.csv", c.filePrefix, monthKey)
	filepath := filepath.Join(c.outputDir, filename)

	// 检查文件是否存在，如果不存在则写入头部
	var writeHeader bool
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		writeHeader = true
	}

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开CSV文件失败: %v", err)
	}

	writer := csv.NewWriter(file)

	// 写入头部（如果是新文件）
	if writeHeader {
		header := []string{
			"id", "url", "title", "subtitle", "raw",
			"publish_date", "edition", "type", "content", "created_at",
		}
		if err := writer.Write(header); err != nil {
			file.Close()
			return nil, fmt.Errorf("写入CSV头部失败: %v", err)
		}
		writer.Flush()
	}

	c.files[monthKey] = file
	c.writers[monthKey] = writer

	return writer, nil
}

// escapeCSVField 转义CSV字段中的特殊字符
func (c *CSVStorage) escapeCSVField(field string) string {
	// 替换换行符为空格
	field = strings.ReplaceAll(field, "\n", " ")
	field = strings.ReplaceAll(field, "\r", " ")

	// 去除多余的空格
	field = strings.TrimSpace(field)

	return field
}

// Close 关闭所有文件
func (c *CSVStorage) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []string

	// 关闭所有写入器和文件
	for monthKey, writer := range c.writers {
		writer.Flush()
		if err := writer.Error(); err != nil {
			errs = append(errs, fmt.Sprintf("刷新写入器 %s 失败: %v", monthKey, err))
		}
	}

	for monthKey, file := range c.files {
		if err := file.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("关闭文件 %s 失败: %v", monthKey, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("关闭CSV存储时出现错误: %s", strings.Join(errs, "; "))
	}

	return nil
}

// GetStorageType 获取存储类型
func (c *CSVStorage) GetStorageType() string {
	return "csv"
}
