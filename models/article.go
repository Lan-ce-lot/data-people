package models

import "time"

// Article 文章数据模型
type Article struct {
	ID          int       `json:"id" db:"id" csv:"id"`                         // id
	URL         string    `json:"url" db:"url" csv:"url"`                     // url 原始链接 也就是请求的链接
	Title       string    `json:"title" db:"title" csv:"title"`               // title 标题
	Subtitle    string    `json:"subtitle" db:"subtitle" csv:"subtitle"`       // subtitle 记者名字/小标题，可能是空的
	Raw         string    `json:"raw" db:"raw" csv:"raw"`                     // raw 特征的内容的全部
	PublishDate time.Time `json:"publish_date" db:"publish_date" csv:"publish_date"` // time 来自特征的内容的时间，如 2025年8月30日
	Edition     string    `json:"edition" db:"edition" csv:"edition"`         // edition 第几版，如第1版
	Type        string    `json:"type" db:"type" csv:"type"`                  // type 类型如要闻，可能是空的
	Content     string    `json:"content" db:"content" csv:"content"`         // content 文章内容
	CreatedAt   time.Time `json:"created_at" db:"created_at" csv:"created_at"` // 创建时间（系统字段）
}

// SearchQuery 搜索查询参数
type SearchQuery struct {
	CDS []SearchCondition `json:"cds"`
	OBS []OrderBy         `json:"obs"`
}

// SearchCondition 搜索条件
type SearchCondition struct {
	Fld string `json:"fld"`
	Cdr string `json:"cdr"`
	Hlt string `json:"hlt"`
	Vlr string `json:"vlr"`
	Qtp string `json:"qtp"`
	Val string `json:"val"`
}

// OrderBy 排序条件
type OrderBy struct {
	Fld string `json:"fld"`
	Drt string `json:"drt"`
}

// APIResponse API响应结构
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Total    int       `json:"total"`
		PageNo   int       `json:"pageNo"`
		PageSize int       `json:"pageSize"`
		Results  []Article `json:"results"`
	} `json:"data"`
}
