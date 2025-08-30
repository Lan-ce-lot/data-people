package crawler

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient HTTP客户端封装
type HTTPClient struct {
	client      *http.Client
	userAgent   string
	timeout     time.Duration
	baseCookies string // 基础Cookie，不包含页码信息
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(timeout time.Duration, userAgent string, baseCookies string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		userAgent:   userAgent,
		timeout:     timeout,
		baseCookies: baseCookies,
	}
}

// GetWithPageInfo 发送带页码信息的GET请求
func (h *HTTPClient) GetWithPageInfo(url string, pageNo, pageSize int) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头，参考实际的curl请求
	req.Header.Set("User-Agent", h.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7,en-GB;q=0.6")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="137", "Chromium";v="137", "Not/A)Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)

	// 添加Cookie支持，包含动态页码信息
	if h.baseCookies != "" {
		cookieWithPageInfo := fmt.Sprintf("%s; pageNo=%d; pageSize=%d", h.baseCookies, pageNo, pageSize)
		req.Header.Set("Cookie", cookieWithPageInfo)
	}

	// 发送请求
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("请求被限流: status=429, 请增加请求间隔")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP请求失败: status=%d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	return body, nil
}

// GetWithRetry 带重试的GET请求
func (h *HTTPClient) GetWithRetry(url string, maxRetries int, retryInterval time.Duration) ([]byte, error) {
	return h.GetWithRetryAndPageInfo(url, maxRetries, retryInterval, 1, 20)
}

// GetWithRetryAndPageInfo 带重试和页码信息的GET请求
func (h *HTTPClient) GetWithRetryAndPageInfo(url string, maxRetries int, retryInterval time.Duration, pageNo, pageSize int) ([]byte, error) {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			// 对于429错误，使用更长的等待时间
			waitTime := retryInterval * time.Duration(1<<uint(i-1))
			if lastErr != nil && containsRateLimit(lastErr.Error()) {
				waitTime = waitTime * 3 // 限流错误等待更长时间
			}
			fmt.Printf("  等待 %v 后重试...\n", waitTime)
			time.Sleep(waitTime)
		}

		body, err := h.GetWithPageInfo(url, pageNo, pageSize)
		if err == nil {
			return body, nil
		}

		lastErr = err
		fmt.Printf("  请求失败 (第%d次): %v\n", i+1, err)

		// 如果是429错误且还有重试机会，继续重试
		if containsRateLimit(err.Error()) && i < maxRetries {
			continue
		}
	}

	return nil, fmt.Errorf("重试%d次后仍然失败: %v", maxRetries, lastErr)
}

// containsRateLimit 检查错误是否包含限流信息
func containsRateLimit(errMsg string) bool {
	return len(errMsg) > 0 && (errMsg[len(errMsg)-3:] == "429" ||
		len(errMsg) > 6 && errMsg[:6] == "请求被限流")
}
