# 人民日报数据爬虫

这是一个用Go语言编写的人民日报网站数据爬虫，用于抓取https://data.people.com.cn/上的文章信息。


## 系统要求

- Go 1.20+
- MySQL 8.0+ (可选)
- 磁盘空间: 建议100GB+
- 内存: 建议1GB+

## 快速开始

### 1. 下载依赖

```bash
go mod tidy
```

### 2. 配置数据库 (可选)

如果使用MySQL存储，需要先创建数据库：

```bash
mysql -u root -p < sql/schema.sql
```

### 3. 修改配置

编辑 `config.yaml` 文件，配置数据库连接信息：

```yaml
storage:
  types: ["csv", "mysql"]  # 可以只使用csv: ["csv"]
  mysql:
    host: "localhost"
    port: 3306
    username: "your_username"
    password: "your_password"
    database: "people_daily"
```

### 4. 运行程序

```bash
	go run main.go crawl --config test_config.yaml
```


## 配置说明

### 主要配置项

```yaml
app:
  name: "人民日报爬虫"
  version: "1.0.0"
  
crawler:
  workers: 5                    # 并发worker数量
  request_interval: 1000ms      # 请求间隔
  timeout: 30s                  # 请求超时
  max_retries: 3               # 最大重试次数
  
date_range:
  start_year: 1949             # 开始年份
  end_year: 2025               # 结束年份
  
storage:
  types: ["csv", "mysql"]      # 存储类型
  csv:
    output_dir: "./data"       # CSV输出目录
    file_prefix: "articles"    # 文件名前缀
```

### 输出文件

- **CSV文件**: 按月份分割，格式为 `articles_YYYYMM.csv`
- **MySQL数据**: 存储在 `articles` 表中

## 数据字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 文章ID |
| title | string | 文章标题 |
| url | string | 文章URL |
| content | string | 文章内容 |
| summary | string | 文章摘要 |
| publish_date | datetime | 发布日期 |
| author | string | 作者 |
| source | string | 来源 |
| keywords | string | 关键词 |
| category | string | 分类 |
| created_at | datetime | 创建时间 |


## 项目结构

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
│   └── parser.go          # 数据解析器
├── storage/
│   ├── interface.go       # 存储接口
│   ├── csv.go             # CSV存储
│   └── mysql.go           # MySQL存储
├── utils/
│   ├── url.go             # URL构建工具
│   └── date.go            # 日期工具
└── sql/
    └── schema.sql         # 数据库表结构
```



## 许可证

本项目仅用于学习和研究目的。请遵守相关法律法规和网站服务条款。
