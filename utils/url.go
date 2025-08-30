package utils

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/Lan-ce-lot/data-people/models"
)

// URLBuilder URL构建器
type URLBuilder struct {
	baseSearchURL string
}

// NewURLBuilder 创建URL构建器
func NewURLBuilder(baseSearchURL string) *URLBuilder {
	if baseSearchURL == "" {
		baseSearchURL = "https://data.people.com.cn/rmrb/pd.html" // 默认值
	}
	return &URLBuilder{
		baseSearchURL: baseSearchURL,
	}
}

// BuildSearchURL 构建搜索URL
func (u *URLBuilder) BuildSearchURL(startDate, endDate time.Time, pageNo, position int) (string, error) {
	// 构建查询条件
	query := models.SearchQuery{
		CDS: []models.SearchCondition{
			{
				Fld: "dataTime.start",
				Cdr: "AND",
				Hlt: "false",
				Vlr: "AND",
				Qtp: "DEF",
				Val: startDate.Format("2006-01-02"),
			},
			{
				Fld: "dataTime.end",
				Cdr: "AND",
				Hlt: "false",
				Vlr: "AND",
				Qtp: "DEF",
				Val: endDate.Format("2006-01-02"),
			},
		},
		OBS: []models.OrderBy{
			{
				Fld: "dataTime",
				Drt: "DESC",
			},
		},
	}

	// 序列化查询条件为JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return "", fmt.Errorf("序列化查询条件失败: %v", err)
	}

	// 构建URL参数，参数顺序按照实际curl请求
	params := url.Values{}
	params.Set("pageNo", fmt.Sprintf("%d", pageNo))
	params.Set("pageSize", "20")
	params.Set("position", fmt.Sprintf("%d", position))
	params.Set("qs", string(queryJSON))
	params.Set("tr", "A")

	// 构建完整URL
	searchURL := fmt.Sprintf("%s?%s", u.baseSearchURL, params.Encode())
	return searchURL, nil
}

// ParseDateRange 解析日期范围，按月份分割
func (u *URLBuilder) ParseDateRange(startYear, endYear int) []DateRange {
	var ranges []DateRange

	for year := startYear; year <= endYear; year++ {
		for month := 1; month <= 12; month++ {
			// 当前月的第一天
			startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

			// 下个月的第一天减去一天，即当前月的最后一天
			var endDate time.Time
			if month == 12 {
				endDate = time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
			} else {
				endDate = time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
			}

			// 如果是未来日期，跳过
			if startDate.After(time.Now()) {
				break
			}

			// 如果结束日期超过当前时间，调整为当前时间
			if endDate.After(time.Now()) {
				endDate = time.Now()
			}

			ranges = append(ranges, DateRange{
				Start: startDate,
				End:   endDate,
			})
		}

		// 如果年份超过当前年份，停止
		if year >= time.Now().Year() {
			break
		}
	}

	return ranges
}

// ParseSpecificDateRange 解析具体日期范围
func (u *URLBuilder) ParseSpecificDateRange(startDateStr, endDateStr string) ([]DateRange, error) {
	// 解析开始日期
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, fmt.Errorf("解析开始日期失败: %v", err)
	}

	// 解析结束日期
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, fmt.Errorf("解析结束日期失败: %v", err)
	}

	// 验证日期范围
	if startDate.After(endDate) {
		return nil, fmt.Errorf("开始日期不能晚于结束日期")
	}

	// 如果日期范围小于等于一个月，直接返回
	if endDate.Sub(startDate) <= 31*24*time.Hour {
		return []DateRange{
			{
				Start: startDate,
				End:   endDate,
			},
		}, nil
	}

	// 如果日期范围较大，按月份分割
	var ranges []DateRange
	current := startDate

	for current.Before(endDate) || current.Equal(endDate) {
		// 当前月的最后一天
		nextMonth := time.Date(current.Year(), current.Month()+1, 1, 0, 0, 0, 0, current.Location())
		monthEnd := nextMonth.AddDate(0, 0, -1)

		// 如果月末超过了结束日期，使用结束日期
		if monthEnd.After(endDate) {
			monthEnd = endDate
		}

		ranges = append(ranges, DateRange{
			Start: current,
			End:   monthEnd,
		})

		// 移动到下个月的第一天
		current = time.Date(current.Year(), current.Month()+1, 1, 0, 0, 0, 0, current.Location())
	}

	return ranges, nil
}

// DateRange 日期范围
type DateRange struct {
	Start time.Time
	End   time.Time
}

// String 返回日期范围的字符串表示
func (dr DateRange) String() string {
	return fmt.Sprintf("%s to %s",
		dr.Start.Format("2006-01-02"),
		dr.End.Format("2006-01-02"))
}

// GetMonthKey 获取月份键值（YYYYMM格式）
func (dr DateRange) GetMonthKey() string {
	return dr.Start.Format("200601")
}
