package MDtransform

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
)

// ConvertMarkdownToHTML 接收一个 Markdown 字符串，返回对应的 HTML 字符串。
func ConvertMarkdownToHTML(markdownText string) (string, error) {
	// 创建一个 goldmark 的默认实例
	md := goldmark.New()

	// 准备输入的字节切片
	var buf bytes.Buffer
	source := []byte(markdownText)

	// 将 Markdown 渲染为 HTML 并写入 buf
	if err := md.Convert(source, &buf); err != nil {
		return "", err // 如果转换出错，则返回错误
	}

	// 将缓冲区的内容转换为字符串并返回
	return buf.String(), nil
}

// 方案2(更推荐)
type TextModel struct {
	ArticleID uint   `json:"articleID"`
	Head      string `json:"head"`
	Body      string `json:"body"`
}

func MdContentTransformation(title string, content string, articleID uint) (list []TextModel) {
	lines := strings.Split(content, "\n")
	var headList []string
	var bodyList []string
	var body string
	headList = append(headList, title)
	var flag bool
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			flag = !flag
		}
		if !flag && strings.HasPrefix(line, "#") {
			// 标题行
			headList = append(headList, getHead(line))
			//if strings.TrimSpace(body) != "" {
			bodyList = append(bodyList, getBody(body))
			//}
			body = ""
			continue
		}
		body += line
	}
	if body != "" {
		bodyList = append(bodyList, getBody(body))
	}

	if len(headList) != len(bodyList) {
		fmt.Println("headList与bodyList 不一致")
		fmt.Printf("%q  %d\n", headList, len(headList))
		fmt.Printf("%q  %d\n", bodyList, len(bodyList))
		return
	}

	for i := 0; i < len(headList); i++ {
		list = append(list, TextModel{
			ArticleID: articleID,
			Head:      headList[i],
			Body:      bodyList[i],
		})
	}

	return

}

func getHead(head string) string {
	s := strings.TrimSpace(strings.Join(strings.Split(head, " ")[1:], " "))
	return s
}

func getBody(body string) string {
	body = strings.TrimSpace(body)
	return body
}
