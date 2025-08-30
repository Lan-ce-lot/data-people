# 人民日报数据爬虫

这是一个用Go语言编写的人民日报网站数据爬虫，用于抓取https://data.people.com.cn/上的文章信息。

## 功能特点

- 🚀 **高性能**: 支持并发抓取，可配置worker数量
- 💾 **双重存储**: 同时支持CSV文件和MySQL数据库存储
- 🔄 **断点续传**: 支持中断后继续抓取
- 🛡️ **稳定可靠**: 内置重试机制和错误处理
- ⚙️ **配置灵活**: 通过YAML文件灵活配置各项参数
- 📊 **监控友好**: 详细的进度统计和日志输出

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
# 编译
go build -o crawler main.go

# 运行
./crawler

# 或者指定配置文件
./crawler -config config.yaml
```

### 5. 查看帮助

```bash
./crawler -help
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

## 性能说明

- **处理速度**: 每秒处理10-20篇文章
- **内存使用**: 通常在500MB以内
- **网络友好**: 内置限流，避免对目标服务器造成压力

## 注意事项

### 合规使用

1. **遵守robots.txt**: 请查看目标网站的robots.txt协议
2. **控制频率**: 默认配置已设置合理的请求间隔
3. **学习用途**: 仅用于学习和研究目的

### 错误处理

- **网络错误**: 自动重试，最多3次
- **解析错误**: 记录日志，跳过当前数据
- **存储错误**: 重试保存，失败后记录错误

### 监控日志

程序运行时会显示详细的进度信息：

```
=== 人民日报爬虫 v1.0.0 ===
配置文件: config.yaml
并发worker数: 5
请求间隔: 1s
存储类型: [csv mysql]

✓ csv存储初始化成功
✓ mysql存储初始化成功
生成了 912 个月份任务 (1949年-2025年)
开始抓取数据...

[1/912] 处理时间段: 1949-01-01 to 1949-01-31
  ✓ 保存 15 篇文章到csv
  ✓ 保存 15 篇文章到mysql
```

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

## 故障排除

### 常见问题

1. **连接超时**
   - 检查网络连接
   - 增加timeout配置
   - 减少并发数量

2. **MySQL连接失败**
   - 检查数据库服务是否启动
   - 确认用户名密码正确
   - 检查数据库是否存在

3. **权限错误**
   - 确保输出目录有写权限
   - 检查MySQL用户权限

### 调试方法

```bash
# 启用详细日志
export GOMAXPROCS=1
./crawler -config config.yaml
```

## 许可证

本项目仅用于学习和研究目的。请遵守相关法律法规和网站服务条款。

## 贡献

欢迎提交Issue和Pull Request来改进这个项目。