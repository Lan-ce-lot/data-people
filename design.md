# 人民日报数据爬虫系统设计文档

## 1. 项目概述

### 1.1 项目背景
人民日报网站（https://data.people.com.cn/）包含大量历史文章数据，需要开发一个高效的爬虫系统来抓取这些文章信息。

### 1.2 项目目标
- 抓取人民日报网站从1949年到2025年的所有文章
- 支持CSV和MySQL双重存储方式
- 实现高效、稳定的并发抓取
- 提供完整的错误处理和重试机制

## 2. 系统架构设计

### 2.1 整体架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Main Entry    │───▶│   Scheduler     │───▶│   Workers       │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Task Queue    │    │   HTTP Client   │
                       │                 │    │                 │
                       └─────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │   Data Parser   │
                                               │                 │
                                               └─────────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │   Storage       │
                                               │   ┌───────────┐ │
                                               │   │    CSV    │ │
                                               │   │   MySQL   │ │
                                               │   └───────────┘ │
                                               └─────────────────┘
```

### 2.2 模块设计

#### 2.2.1 目录结构
```
data-people/
├── main.go                 # 程序入口
├── config/
│   ├── config.go          # 配置管理
│   └── config.yaml        # 配置文件
├── models/
│   ├── article.go         # 文章数据模型
│   └── task.go            # 任务模型
├── crawler/
│   ├── client.go          # HTTP客户端
│   ├── scheduler.go       # 任务调度器
│   ├── worker.go          # 工作协程
│   └── parser.go          # 数据解析器
├── storage/
│   ├── interface.go       # 存储接口定义
│   ├── csv.go             # CSV存储实现
│   ├── mysql.go           # MySQL存储实现
│   └── factory.go         # 存储工厂
├── utils/
│   ├── date.go            # 日期工具
│   ├── url.go             # URL构建工具
│   └── logger.go          # 日志工具
└── sql/
    └── schema.sql         # 数据库表结构
```

## 3. 核心模块详细设计

### 3.1 数据模型设计

#### 3.1.1 文章模型 (Article)
```go
type Article struct {
    ID          int       `json:"id" db:"id" csv:"id"`
    Title       string    `json:"title" db:"title" csv:"title"`
    URL         string    `json:"url" db:"url" csv:"url"`
    Content     string    `json:"content" db:"content" csv:"content"`
    Summary     string    `json:"summary" db:"summary" csv:"summary"`
    PublishDate time.Time `json:"publish_date" db:"publish_date" csv:"publish_date"`
    Author      string    `json:"author" db:"author" csv:"author"`
    Source      string    `json:"source" db:"source" csv:"source"`
    Keywords    string    `json:"keywords" db:"keywords" csv:"keywords"`
    Category    string    `json:"category" db:"category" csv:"category"`
    CreatedAt   time.Time `json:"created_at" db:"created_at" csv:"created_at"`
}
```

#### 3.1.2 任务模型 (Task)
```go
type Task struct {
    ID        string    `json:"id"`
    StartDate time.Time `json:"start_date"`
    EndDate   time.Time `json:"end_date"`
    PageNo    int       `json:"page_no"`
    Status    string    `json:"status"` // pending, running, completed, failed
    RetryCount int      `json:"retry_count"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 3.2 API接口分析

#### 3.2.1 搜索API
- **URL**: `https://data.people.com.cn/rmrb/pd.html`
- **方法**: GET
- **参数**:
  - `qs`: JSON查询条件（URL编码）
  - `tr`: 固定值 "A"
  - `pageNo`: 页码，从1开始
  - `pageSize`: 每页大小，固定20
  - `position`: 位置偏移，从0开始，每页递增20

#### 3.2.2 查询条件JSON结构
```json
{
  "cds": [
    {
      "fld": "dataTime.start",
      "cdr": "AND",
      "hlt": "false",
      "vlr": "AND",
      "qtp": "DEF",
      "val": "开始日期"
    },
    {
      "fld": "dataTime.end",
      "cdr": "AND",
      "hlt": "false",
      "vlr": "AND",
      "qtp": "DEF",
      "val": "结束日期"
    }
  ],
  "obs": [
    {
      "fld": "dataTime",
      "drt": "DESC"
    }
  ]
}
```

### 3.3 存储系统设计

#### 3.3.1 存储接口
```go
type Storage interface {
    Save(articles []Article) error
    SaveBatch(articles []Article) error
    Close() error
    Init() error
}
```

#### 3.3.2 CSV存储
- **文件格式**: UTF-8编码的CSV文件
- **文件命名**: `articles_YYYYMM.csv`
- **字段分隔符**: 逗号
- **特殊字符处理**: 转义双引号和换行符

#### 3.3.3 MySQL存储
- **表结构**:
```sql
CREATE TABLE articles (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    url VARCHAR(1000) NOT NULL UNIQUE,
    content LONGTEXT,
    summary TEXT,
    publish_date DATETIME NOT NULL,
    author VARCHAR(200),
    source VARCHAR(200),
    keywords VARCHAR(500),
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_publish_date (publish_date),
    INDEX idx_author (author),
    INDEX idx_category (category)
);
```

### 3.4 并发控制设计

#### 3.4.1 任务调度策略
- **时间分片**: 按月份划分任务（1949-01 到 2025-12）
- **并发控制**: 可配置的Worker数量
- **限流机制**: 请求间隔控制，避免对服务器造成压力
- **错误重试**: 指数退避重试策略

#### 3.4.2 Worker Pool模式
```go
type WorkerPool struct {
    workers    int
    taskQueue  chan Task
    resultChan chan []Article
    errChan    chan error
    wg         sync.WaitGroup
}
```

## 4. 配置管理

### 4.1 配置文件结构 (config.yaml)
```yaml
app:
  name: "人民日报爬虫"
  version: "1.0.0"
  
crawler:
  workers: 5                    # 并发worker数量
  request_interval: 1000        # 请求间隔(毫秒)
  timeout: 30                   # 请求超时(秒)
  max_retries: 3               # 最大重试次数
  user_agent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
  
date_range:
  start_year: 1949
  end_year: 2025
  
storage:
  types: ["csv", "mysql"]      # 启用的存储类型
  csv:
    output_dir: "./data"       # CSV文件输出目录
    file_prefix: "articles"    # 文件名前缀
  mysql:
    host: "localhost"
    port: 3306
    username: "root"
    password: "password"
    database: "people_daily"
    charset: "utf8mb4"
    max_open_conns: 10
    max_idle_conns: 5
    
logging:
  level: "info"                # debug, info, warn, error
  file: "./logs/crawler.log"
  max_size: 100               # MB
  max_backups: 3
  max_age: 28                 # days
```

## 5. 错误处理和监控

### 5.1 错误处理策略
- **网络错误**: 自动重试，指数退避
- **解析错误**: 记录错误日志，跳过当前数据
- **存储错误**: 重试机制，失败后记录到错误文件
- **限流错误**: 动态调整请求间隔

### 5.2 监控指标
- 抓取进度统计
- 成功/失败请求数
- 平均响应时间
- 存储成功率
- 错误类型分布

### 5.3 日志设计
```
[2025-08-30 10:30:15] [INFO] 开始抓取任务: 1949-01-01 to 1949-01-31
[2025-08-30 10:30:16] [INFO] 发送请求: pageNo=1, position=0
[2025-08-30 10:30:17] [INFO] 解析到 20 篇文章
[2025-08-30 10:30:18] [INFO] 保存到CSV成功: 20 篇文章
[2025-08-30 10:30:19] [INFO] 保存到MySQL成功: 20 篇文章
[2025-08-30 10:30:20] [ERROR] 请求失败，开始重试: 第1次
```

## 6. 性能优化

### 6.1 性能目标
- 每秒处理 10-20 篇文章
- 内存使用控制在 500MB 以内
- 支持 7x24 小时稳定运行

### 6.2 优化策略
- **连接池**: HTTP连接复用
- **批量处理**: 批量写入数据库
- **内存管理**: 及时释放大对象
- **缓存机制**: 避免重复抓取

## 7. 部署和运维

### 7.1 部署要求
- Go 1.20+
- MySQL 8.0+
- 磁盘空间: 预估100GB+
- 内存: 1GB+

### 7.2 运行命令
```bash
# 编译
go build -o crawler main.go

# 运行
./crawler -config config.yaml

# 查看帮助
./crawler -help
```

### 7.3 数据备份
- 定期备份MySQL数据
- CSV文件分目录存储
- 重要日志文件保留

## 8. 风险和限制

### 8.1 技术风险
- 目标网站反爬虫策略变化
- API接口结构调整
- 网络不稳定导致数据丢失

### 8.2 法律合规
- 遵守robots.txt协议
- 控制访问频率，避免对服务器造成负担
- 仅用于学习和研究目的

### 8.3 数据质量
- 网页内容解析可能不完整
- 历史数据可能存在缺失
- 编码问题导致乱码

## 9. 后续扩展

### 9.1 功能扩展
- 支持增量更新
- 添加全文搜索功能
- 数据分析和可视化
- 支持其他新闻网站

### 9.2 技术改进
- 使用消息队列优化任务调度
- 添加分布式部署支持
- 集成机器学习进行内容分类



---

## xpath

我们现在要处理html页面的内容了，基于xpath我们发现以下特征，
/html[1]/body[1]/div[1]/div[1]/div[2]/div[1]/div[3]是文章的所有内容的div
其中有部分文章
/html[1]/body[1]/div[1]/div[1]/div[2]/div[1]/div[1]是标题， title

/html[1]/body[1]/div[1]/div[1]/div[2]/div[1]/div[2]也是标题/记者名字，subtitle
但有些
/html[1]/body[1]/div[1]/div[1]/div[2]/div[1]/div[2]直接就是,这种格式的存在特征的内容 `【字号：加大还原减小】`
```
【人民日报2025年8月30日
								第1版
							要闻】
						
						
							【字号：加大还原减小】

```
现在的方案是识别出这种内容，
假设这种是 n
/html[1]/body[1]/div[1]/div[1]/div[2]/div[1]/div[n]
可以保证 
/html[1]/body[1]/div[1]/div[1]/div[2]/div[1]/div[n + 1]就是中文

id 
url  原始链接 也就是请求的链接
title 标题 
subtitle 记者名字/小标题 ，可能是空的
raw 特征的内容的全部
time 来自特征的内容的时间，如 2025年8月30日
edition  第几版，如第1版
type 类型 如要闻，可能是空的
content 文章内容



