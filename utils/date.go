package utils

import (
	"time"
)

// DateHelper 日期辅助工具
type DateHelper struct{}

// NewDateHelper 创建日期辅助工具
func NewDateHelper() *DateHelper {
	return &DateHelper{}
}

// FormatTime 格式化时间
func (d *DateHelper) FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseTime 解析时间字符串
func (d *DateHelper) ParseTime(timeStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, nil
}

// GetMonthStart 获取月份开始时间
func (d *DateHelper) GetMonthStart(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}

// GetMonthEnd 获取月份结束时间
func (d *DateHelper) GetMonthEnd(year int, month time.Month) time.Time {
	nextMonth := d.GetMonthStart(year, month).AddDate(0, 1, 0)
	return nextMonth.AddDate(0, 0, -1).Add(time.Hour*23 + time.Minute*59 + time.Second*59)
}

// IsValidDateRange 检查日期范围是否有效
func (d *DateHelper) IsValidDateRange(start, end time.Time) bool {
	return start.Before(end) && !start.After(time.Now())
}

// GetDateRangeInDays 获取日期范围的天数
func (d *DateHelper) GetDateRangeInDays(start, end time.Time) int {
	duration := end.Sub(start)
	return int(duration.Hours() / 24)
}
