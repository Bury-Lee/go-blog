package xss_filter

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

type AdvancedXSSFilter struct {
	policy *bluemonday.Policy
}

/*
使用示例:

	func main() {
		// 初始化过滤器
		filter := NewXSSFilter()

		testCases := []string{
			`<script>alert('XSS')</script>`,
			`<img src="x" onerror="alert('XSS')">`,
			`<a href="javascript:alert('XSS')">Click me</a>`,
			`<p style="color:red">Safe content</p>`,
			`<img src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA..." alt="image">`,
			`Normal text with <strong>safe</strong> HTML`,
		}

		for _, testCase := range testCases {
			sanitized := filter.Sanitize(testCase)
			fmt.Printf("Input:  %s\n", testCase)
			fmt.Printf("Output: %s\n\n", sanitized)
		}
	}
*/
func NewXSSFilter() *AdvancedXSSFilter {
	policy := bluemonday.NewPolicy()
	policy.AllowElements(
		"p", "br", "strong", "em",
		"h1", "h2", "h3", "h4",
		"ul", "ol", "li", "a",
		"img", "div", "span",
	)
	policy.AllowAttrs("href", "title").OnElements("a")
	policy.AllowAttrs("src", "alt", "title").OnElements("img")
	policy.AllowAttrs("class").Matching(regexp.MustCompile(`^[a-zA-Z0-9_\-\s]+$`)).OnElements("div", "span")
	policy.AllowStandardURLs()
	policy.AllowURLSchemes("http", "https", "mailto", "tel")
	policy.AllowDataURIImages()
	policy.RequireNoFollowOnLinks(true)
	policy.RequireNoReferrerOnLinks(true)
	policy.AddTargetBlankToFullyQualifiedLinks(true)

	return &AdvancedXSSFilter{policy: policy}
}

func (f *AdvancedXSSFilter) Sanitize(input string) string {
	if f == nil || f.policy == nil {
		return ""
	}
	return f.policy.Sanitize(input)
}
