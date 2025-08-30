package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Lan-ce-lot/data-people/models"
	_ "github.com/go-sql-driver/mysql"
)

// MySQLStorage MySQL存储实现
type MySQLStorage struct {
	db       *sql.DB
	dsn      string
	prepared map[string]*sql.Stmt
}

// NewMySQLStorage 创建MySQL存储实例
func NewMySQLStorage(host string, port int, username, password, database, charset string) *MySQLStorage {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		username, password, host, port, database, charset)

	return &MySQLStorage{
		dsn:      dsn,
		prepared: make(map[string]*sql.Stmt),
	}
}

// Init 初始化MySQL连接
func (m *MySQLStorage) Init() error {
	db, err := sql.Open("mysql", m.dsn)
	if err != nil {
		return fmt.Errorf("连接MySQL失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("MySQL连接测试失败: %v", err)
	}

	m.db = db

	// 设置连接池参数
	m.db.SetMaxOpenConns(10)
	m.db.SetMaxIdleConns(5)
	m.db.SetConnMaxLifetime(time.Hour)

	// 创建表（如果不存在）
	if err := m.createTable(); err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	// 预编译SQL语句
	if err := m.prepareSQLStatements(); err != nil {
		return fmt.Errorf("预编译SQL语句失败: %v", err)
	}

	return nil
}

// createTable 创建文章表
func (m *MySQLStorage) createTable() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS articles (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(500) NOT NULL,
		url VARCHAR(1000) NOT NULL,
		content LONGTEXT,
		summary TEXT,
		publish_date DATETIME NOT NULL,
		author VARCHAR(200),
		source VARCHAR(200),
		keywords VARCHAR(500),
		category VARCHAR(100),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE KEY uk_url (url),
		INDEX idx_publish_date (publish_date),
		INDEX idx_author (author),
		INDEX idx_category (category)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	_, err := m.db.Exec(createTableSQL)
	return err
}

// prepareSQLStatements 预编译SQL语句
func (m *MySQLStorage) prepareSQLStatements() error {
	// 插入单条记录的SQL
	insertSQL := `
	INSERT INTO articles (url, title, subtitle, raw, publish_date, edition, type, content, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		title = VALUES(title),
		subtitle = VALUES(subtitle),
		raw = VALUES(raw),
		publish_date = VALUES(publish_date),
		edition = VALUES(edition),
		type = VALUES(type),
		content = VALUES(content)
	`

	stmt, err := m.db.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("预编译插入语句失败: %v", err)
	}
	m.prepared["insert"] = stmt

	return nil
}

// Save 保存单个文章
func (m *MySQLStorage) Save(article *models.Article) error {
	return m.SaveBatch([]*models.Article{article})
}

// SaveBatch 批量保存文章
func (m *MySQLStorage) SaveBatch(articles []*models.Article) error {
	if len(articles) == 0 {
		return nil
	}

	// 开始事务
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 准备批量插入语句
	stmt := tx.Stmt(m.prepared["insert"])
	defer stmt.Close()

	// 批量插入
	for _, article := range articles {
		_, err := stmt.Exec(
			article.URL,
			article.Title,
			article.Subtitle,
			article.Raw,
			article.PublishDate,
			article.Edition,
			article.Type,
			article.Content,
			article.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("插入文章失败 [%s]: %v", article.URL, err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// Close 关闭数据库连接
func (m *MySQLStorage) Close() error {
	var errs []string

	// 关闭预编译语句
	for name, stmt := range m.prepared {
		if err := stmt.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("关闭预编译语句 %s 失败: %v", name, err))
		}
	}

	// 关闭数据库连接
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("关闭数据库连接失败: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("关闭MySQL存储时出现错误: %s", strings.Join(errs, "; "))
	}

	return nil
}

// GetStorageType 获取存储类型
func (m *MySQLStorage) GetStorageType() string {
	return "mysql"
}

// GetStats 获取存储统计信息
func (m *MySQLStorage) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取文章总数
	var totalCount int
	err := m.db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("获取文章总数失败: %v", err)
	}
	stats["total_articles"] = totalCount

	// 获取最早和最晚的文章日期
	var earliestDate, latestDate sql.NullTime
	err = m.db.QueryRow("SELECT MIN(publish_date), MAX(publish_date) FROM articles").Scan(&earliestDate, &latestDate)
	if err != nil {
		return nil, fmt.Errorf("获取日期范围失败: %v", err)
	}
	if earliestDate.Valid {
		stats["earliest_date"] = earliestDate.Time.Format("2006-01-02")
	}
	if latestDate.Valid {
		stats["latest_date"] = latestDate.Time.Format("2006-01-02")
	}

	return stats, nil
}
