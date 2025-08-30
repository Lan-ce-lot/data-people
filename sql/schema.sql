-- 人民日报文章数据表结构

-- 创建数据库
CREATE DATABASE IF NOT EXISTS people_daily 
DEFAULT CHARACTER SET utf8mb4 
DEFAULT COLLATE utf8mb4_unicode_ci;

USE people_daily;

-- 创建文章表
CREATE TABLE IF NOT EXISTS articles (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'id - 文章ID',
    url VARCHAR(1000) NOT NULL COMMENT 'url - 原始链接',
    title VARCHAR(500) NOT NULL COMMENT 'title - 标题',
    subtitle VARCHAR(500) COMMENT 'subtitle - 记者名字/小标题，可能是空的',
    raw TEXT COMMENT 'raw - 特征的内容的全部',
    publish_date DATETIME COMMENT 'publish_date - 来自特征的内容的时间，如 2025年8月30日',
    edition VARCHAR(50) COMMENT 'edition - 第几版，如第1版',
    type VARCHAR(100) COMMENT 'type - 类型如要闻，可能是空的',
    content LONGTEXT COMMENT 'content - 文章内容',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间（系统字段）',
    
    -- 索引
    UNIQUE KEY uk_url (url) COMMENT 'URL唯一索引',
    INDEX idx_publish_date (publish_date) COMMENT '发布日期索引',
    INDEX idx_title (title) COMMENT '标题索引',
    INDEX idx_edition (edition) COMMENT '版次索引',
    INDEX idx_type (type) COMMENT '类型索引',
    INDEX idx_created_at (created_at) COMMENT '创建时间索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='人民日报文章表';

-- 创建统计表（可选）
CREATE TABLE IF NOT EXISTS crawl_stats (
    id INT AUTO_INCREMENT PRIMARY KEY,
    crawl_date DATE NOT NULL COMMENT '爬取日期',
    articles_count INT DEFAULT 0 COMMENT '文章数量',
    success_count INT DEFAULT 0 COMMENT '成功数量',
    failed_count INT DEFAULT 0 COMMENT '失败数量',
    start_time DATETIME COMMENT '开始时间',
    end_time DATETIME COMMENT '结束时间',
    duration_seconds INT COMMENT '耗时(秒)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_crawl_date (crawl_date) COMMENT '爬取日期唯一索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='爬取统计表';

-- 插入示例查询
-- 按年份统计文章数量
-- SELECT YEAR(publish_date) as year, COUNT(*) as count 
-- FROM articles 
-- GROUP BY YEAR(publish_date) 
-- ORDER BY year;

-- 按月份统计文章数量
-- SELECT DATE_FORMAT(publish_date, '%Y-%m') as month, COUNT(*) as count 
-- FROM articles 
-- GROUP BY DATE_FORMAT(publish_date, '%Y-%m') 
-- ORDER BY month;

-- 查找最新文章
-- SELECT title, publish_date, author 
-- FROM articles 
-- ORDER BY publish_date DESC 
-- LIMIT 10;
