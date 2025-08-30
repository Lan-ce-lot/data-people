package crawler

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/Lan-ce-lot/data-people/models"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Parser 数据解析器
type Parser struct {
	httpClient *HTTPClient
}

// NewParser 创建数据解析器
func NewParser(httpClient *HTTPClient) *Parser {
	return &Parser{
		httpClient: httpClient,
	}
}

// ParseSearchResponse 解析搜索响应
func (p *Parser) ParseSearchResponse(responseBody []byte, searchURL string) (*models.APIResponse, error) {
	// 首先检查是否是HTML响应
	html := string(responseBody)
	if strings.Contains(html, "<html") || strings.Contains(html, "<!DOCTYPE") {
		// 这是HTML响应，需要解析HTML中的文章列表
		log.Printf("收到HTML搜索结果页面，长度: %d 字节", len(responseBody))
		return p.parseHTMLSearchResults(html, searchURL)
	}

	// 尝试解析JSON响应
	var response models.APIResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		// 如果不是标准JSON，尝试从HTML中提取JSON数据
		return p.parseHTMLResponseData(responseBody, searchURL)
	}

	return &response, nil
}

// parseHTMLSearchResults 解析HTML响应 - 实际上是文章页面
func (p *Parser) parseHTMLSearchResults(html string, searchURL string) (*models.APIResponse, error) {
	// 直接解析这个HTML页面作为单篇文章
	article, err := p.ParseHTMLStructure(html)
	if err != nil {
		log.Printf("解析文章HTML失败: %v", err)
		// 返回空结果而不是错误，保持程序继续运行
		return p.createEmptyResponse(), nil
	}

	var articles []models.Article
	if article != nil {
		article.URL = searchURL
		log.Printf("从HTML页面解析到文章: %s", article.Title)
		articles = append(articles, *article)
	}

	response := &models.APIResponse{
		Code:    200,
		Message: "HTML article parsed",
		Data: struct {
			Total    int              `json:"total"`
			PageNo   int              `json:"pageNo"`
			PageSize int              `json:"pageSize"`
			Results  []models.Article `json:"results"`
		}{
			Total:    len(articles),
			PageNo:   1,
			PageSize: len(articles),
			Results:  articles,
		},
	}
	return response, nil
}

// createEmptyResponse 创建空响应
func (p *Parser) createEmptyResponse() *models.APIResponse {
	return &models.APIResponse{
		Code:    200,
		Message: "no article found in HTML",
		Data: struct {
			Total    int              `json:"total"`
			PageNo   int              `json:"pageNo"`
			PageSize int              `json:"pageSize"`
			Results  []models.Article `json:"results"`
		}{
			Total:    0,
			PageNo:   1,
			PageSize: 20,
			Results:  []models.Article{},
		},
	}
}

// parseHTMLResponseData 从HTML响应中解析JSON数据
func (p *Parser) parseHTMLResponseData(htmlBody []byte, searchURL string) (*models.APIResponse, error) {
	html := string(htmlBody)

	// 查找JSON数据的正则表达式模式
	patterns := []string{
		`window\.searchResult\s*=\s*({.*?});`,
		`var\s+searchResult\s*=\s*({.*?});`,
		`"data":\s*({.*?"results":\s*\[.*?\].*?})`,
		`<script[^>]*>.*?var\s+data\s*=\s*({.*?});.*?</script>`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?s)` + pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			var response models.APIResponse
			if err := json.Unmarshal([]byte(matches[1]), &response.Data); err == nil {
				response.Code = 200
				response.Message = "success"
				return &response, nil
			}
		}
	}

	// 如果找不到JSON数据，返回空结果而不是错误
	log.Printf("在HTML中未找到JSON数据，返回空结果!")
	response := &models.APIResponse{
		Code:    200,
		Message: "no data found in HTML",
		Data: struct {
			Total    int              `json:"total"`
			PageNo   int              `json:"pageNo"`
			PageSize int              `json:"pageSize"`
			Results  []models.Article `json:"results"`
		}{
			Total:    0,
			PageNo:   1,
			PageSize: 20,
			Results:  []models.Article{},
		},
	}
	return response, nil
}

// ParseHTMLStructure 基于xpath特征解析HTML结构 (公开方法)
func (p *Parser) ParseHTMLStructure(htmlContent string) (*models.Article, error) {
	article := &models.Article{}

	// 解析HTML文档
	doc, err := htmlquery.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %v", err)
	}

	// 基于设计文档的xpath特征进行解析
	// /html[1]/body[1]/div[1]/div[1]/div[2]/div[1] 是主容器

	// 1. 提取标题 - /html[1]/body[1]/div[1]/div[1]/div[2]/div[1]/div[1]
	titleNode := htmlquery.FindOne(doc, "//html/body/div[1]/div[1]/div[2]/div[1]/div[1]")
	if titleNode != nil {
		article.Title = strings.TrimSpace(htmlquery.InnerText(titleNode))
	}

	// 2. 查找特征内容的位置
	// 遍历 /html[1]/body[1]/div[1]/div[1]/div[2]/div[1] 下的所有div
	containerPath := "//html/body/div[1]/div[1]/div[2]/div[1]"
	containerNode := htmlquery.FindOne(doc, containerPath)
	if containerNode == nil {
		log.Printf("警告: 未找到主容器div，返回空文章")
		// 返回空文章而不是错误，避免程序崩溃
		article.CreatedAt = time.Now()
		return article, nil
	}

	// 获取容器下的所有直接子div
	childDivs := htmlquery.Find(containerNode, "./div")
	if len(childDivs) == 0 {
		log.Printf("警告: 主容器下未找到子div，返回空文章")
		// 返回空文章而不是错误，避免程序崩溃
		article.CreatedAt = time.Now()
		return article, nil
	}

	// 3. 按照设计文档的逻辑解析各字段
	p.parseArticleFieldsWithXPath(childDivs, article)

	// 4. 设置默认值
	if article.CreatedAt.IsZero() {
		article.CreatedAt = time.Now()
	}

	return article, nil
}

// parseArticleFieldsWithXPath 使用xpath解析文章字段
func (p *Parser) parseArticleFieldsWithXPath(childDivs []*html.Node, article *models.Article) {
	if len(childDivs) == 0 {
		return
	}

	// 根据设计文档：
	// div[1] 是标题 title (已在上面解析)
	// div[2] 可能是副标题 subtitle 或者直接是特征内容
	// 特征内容包含【字号：加大还原减小】或【人民日报YYYY年MM月DD日 第X版 类型】
	// 特征内容后一个div是正文

	// 1. 第一个div是标题 (已在ParseHTMLStructure中处理)
	if len(childDivs) > 0 && article.Title == "" {
		article.Title = strings.TrimSpace(htmlquery.InnerText(childDivs[0]))
	}

	// 2. 查找特征内容的位置
	featureIndex := -1
	for i := 1; i < len(childDivs); i++ {
		divText := strings.TrimSpace(htmlquery.InnerText(childDivs[i]))
		if p.isFeatureContent(divText) {
			featureIndex = i
			article.Raw = divText

			// 解析特征内容中的时间、版次、类型
			p.parseFeatureFields(divText, article)
			break
		}
	}

	// 3. 处理subtitle和content
	if featureIndex > 1 {
		// 特征内容前面有其他div，第二个div是subtitle
		article.Subtitle = strings.TrimSpace(htmlquery.InnerText(childDivs[1]))

		// 特征内容后面的所有div是正文
		var contentParts []string
		for i := featureIndex + 1; i < len(childDivs); i++ {
			cleanContent := strings.TrimSpace(htmlquery.InnerText(childDivs[i]))
			if cleanContent != "" {
				contentParts = append(contentParts, cleanContent)
			}
		}
		article.Content = strings.Join(contentParts, "\n\n")

	} else if featureIndex == 1 {
		// 第二个div就是特征内容，没有subtitle

		// 特征内容后面的所有div是正文
		var contentParts []string
		for i := featureIndex + 1; i < len(childDivs); i++ {
			cleanContent := strings.TrimSpace(htmlquery.InnerText(childDivs[i]))
			if cleanContent != "" {
				contentParts = append(contentParts, cleanContent)
			}
		}
		article.Content = strings.Join(contentParts, "\n\n")

	} else {
		// 没有找到特征内容，按默认方式处理
		if len(childDivs) > 1 {
			article.Subtitle = strings.TrimSpace(htmlquery.InnerText(childDivs[1]))
		}

		var contentParts []string
		startIndex := 2
		if article.Subtitle == "" {
			startIndex = 1
		}

		for i := startIndex; i < len(childDivs); i++ {
			cleanContent := strings.TrimSpace(htmlquery.InnerText(childDivs[i]))
			if cleanContent != "" {
				contentParts = append(contentParts, cleanContent)
			}
		}
		article.Content = strings.Join(contentParts, "\n\n")
	}
}

// isFeatureContent 判断是否为特征内容
func (p *Parser) isFeatureContent(content string) bool {
	// 特征内容包含以下特征之一：
	// 1. 【字号：加大还原减小】
	// 2. 【人民日报YYYY年MM月DD日 第X版 类型】
	// 3. 人民日报 + 日期 + 版次

	return strings.Contains(content, "【字号：") ||
		strings.Contains(content, "人民日报") &&
			(strings.Contains(content, "年") && strings.Contains(content, "月") && strings.Contains(content, "日")) ||
		(strings.Contains(content, "第") && strings.Contains(content, "版"))
}

// parseFeatureFields 从特征内容中解析时间、版次、类型
func (p *Parser) parseFeatureFields(raw string, article *models.Article) {
	// 1. 解析时间 (如: 2025年8月30日)
	timePattern := `(\d{4})年(\d{1,2})月(\d{1,2})日`
	timeRe := regexp.MustCompile(timePattern)
	if timeMatch := timeRe.FindStringSubmatch(raw); len(timeMatch) > 3 {
		timeStr := fmt.Sprintf("%s年%s月%s日", timeMatch[1], timeMatch[2], timeMatch[3])
		if publishTime, err := time.Parse("2006年1月2日", timeStr); err == nil {
			article.PublishDate = publishTime
		}
	}

	// 2. 解析版次 (如: 第1版)
	editionPattern := `第(\d+)版`
	editionRe := regexp.MustCompile(editionPattern)
	if editionMatch := editionRe.FindStringSubmatch(raw); len(editionMatch) > 1 {
		article.Edition = "第" + editionMatch[1] + "版"
	}

	// 3. 解析类型 (如: 要闻)
	// 查找版次后面的内容，或者】后面的非空白内容
	typePatterns := []string{
		`第\d+版[^】]*?([^】\s\n]+)`, // 版次后面的内容
		`】\s*([^】\s\n【]+)`,       // 】后面的内容
	}

	for _, pattern := range typePatterns {
		typeRe := regexp.MustCompile(pattern)
		if typeMatch := typeRe.FindStringSubmatch(raw); len(typeMatch) > 1 {
			typeStr := strings.TrimSpace(typeMatch[1])
			if typeStr != "" && !strings.Contains(typeStr, "字号") {
				article.Type = typeStr
				break
			}
		}
	}
}
