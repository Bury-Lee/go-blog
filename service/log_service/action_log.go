// service/log_service/action_log.go
package log_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/utils/ip"
	jwts "StarDreamerCyberNook/utils/jwts"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ActionLog struct { //操作导致异常后使用这个进行记录
	c     *gin.Context
	level enum.LogLevel
	title string

	requestBody  []byte //请求体
	responseBody []byte //响应体

	log                *models.LogModel // 日志模型
	showRequestHeader  bool             // 是否展示请求头
	showResponseHeader bool             // 是否展示响应头
	showRequest        bool             // 是否展示请求体
	showResponse       bool             // 是否展示响应体
	itemList           []string         // 项列表,通过结构体的set系列方法调用来实现日志流在服务中添加项的功能
	ResponseHeader     http.Header
	isMiddlewareSave   bool //是否在中间件中保存
}

func (this *ActionLog) ShowRequest() { //是否把请求头加入到日志项中
	this.showRequest = true
}

func (this *ActionLog) ShowResponse() { //是否把回应体加入到日志项
	this.showResponse = true
}

func (this *ActionLog) ShowRequestHeader() { //是否把请求头加入到日志项
	this.showRequestHeader = true
}

func (this *ActionLog) ShowResponseHeader() { //是否把回应头加入到日志项
	this.showResponseHeader = true
}

func (this *ActionLog) SetLevel(level enum.LogLevel) { //设置日志项等级
	this.level = level
}

func (this *ActionLog) setItem(lable string, value any, LogLevelType enum.LogLevel) { //设置项,是各等级设置项的具体实现
	var v string
	t := reflect.TypeOf(value)
	switch t.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array: //如果是复杂类型就格式化输出
		bytedata, _ := json.Marshal(value)
		v = string(bytedata)
	default:
		v = fmt.Sprintf("%v", value)
	}

	this.itemList = append(this.itemList, fmt.Sprintf("<div class=\"log_item %s\"><div class=\"log_item_label\">%s</div><div class=\"log_item_content\">%s</div></div>",
		LogLevelType.String(),
		lable,
		v)) //存储等级,标签,内容
}
func (this *ActionLog) SetItem(lable string, value any) { //默认等级为info
	this.setItem(lable, value, enum.LogInfoLevel)
}
func (this *ActionLog) SetItemInfo(lable string, value any) { //设置并加入info等级的项
	this.setItem(lable, value, enum.LogInfoLevel)
}
func (this *ActionLog) SetItemWarn(lable string, value any) { //设置warn等级的项
	this.setItem(lable, value, enum.LogWarnLevel)
}
func (this *ActionLog) SetItemError(lable string, value any) { //设置error等级的项
	this.setItem(lable, value, enum.LogErrorLevel)
}

func (this *ActionLog) SetError(label string, err error) { //设置error等级的项,并追踪错误栈
	var msg = make([]byte, 1024)
	n := runtime.Stack(msg, false)
	logrus.Errorf("出错:%s", err.Error())
	this.itemList = append(this.itemList, fmt.Sprintf("<div class=\"log_error\"><div class=\"line\"><div class=\"label\">%s</div><div class=\"value\">%s</div><div class=\"type\">%t</div></div><div class=\"stack\">%+v</div></div>",
		label, err, err, msg[:n]))
	//TODO:记录err
}

func (this *ActionLog) SetTitle(title string) { //设置标题
	this.title = title
}

func (this *ActionLog) SetLink(label string, href string) { //设置链接项
	this.itemList = append(this.itemList, fmt.Sprintf("<div class=\"log_item Link\"><div class=\"log_item_label\">%s</div><div class=\"Log_item_content\"><a href=\"%s\" target=\"blank\">%s</a></div></div>", label, href, href))
}

func (this *ActionLog) SetImage(src string) { //设置图片项
	this.itemList = append(this.itemList, fmt.Sprintf("<div class=\"log_Image\"><img src=\"%s\" alt=\"\"/></div>", src))
}

func (this *ActionLog) SetRequest(c *gin.Context) { //设置请求项
	// 1. 读取请求体并打印
	byteData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Errorf("failed to read request body: %s", err.Error())
	}
	// fmt.Println("body: ", string(byteData))测试
	// 重新设置请求体（因为 ReadAll 已经读完了）
	// 注意：这里必须把原来的 body 重新放回去，否则后续 handler 无法读取
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData))
	this.requestBody = byteData
}
func (this *ActionLog) SetResponse(data []byte) { //设置响应项
	this.responseBody = data
}
func (this *ActionLog) SetResponseHeader(head http.Header) { //设置响应头项
	this.ResponseHeader = head
}

func (this *ActionLog) MiddlewareSave() { //专门为中间件使用的版本
	_saveLog, _ := this.c.Get("saveLog")
	saveLog, _ := _saveLog.(bool)
	if !saveLog {
		return
	}

	if this.log == nil {
		this.isMiddlewareSave = true
		this.Save() //如果日志为空,则保存日志并返回
		return
	} //否则说明已经在视图中保存过,这里就是更新的逻辑
	if this.showResponseHeader {
		byteData, _ := json.Marshal(this.ResponseHeader)
		this.itemList = append(this.itemList, fmt.Sprintf("<div class=\"log_response_header\"><pre class=\"log_json_body\">%s</pre></div>", string(byteData))) //响应头
	}

	if this.showResponse {
		this.itemList = append(this.itemList, fmt.Sprintf("<div class=\"log_response\"><pre class=\"log_json_body\">%s</pre></div>", string(this.responseBody)))
	}
	this.Save()

}

func (this *ActionLog) Save() (id uint) { //保存日志
	//注1:save只能在中间件中调用一次
	//注2:违反注1的话(微服务架构时),就需要在视图中调用save函数并且返回日志id

	if this.log != nil {
		//注2:因为这个逻辑,虽然难以解释,但是绝对会出现问题,解决方法见注1

		content := this.log.Content + "\n" + strings.Join(this.itemList, "\n")
		//!=nil说明此处之前已经保存过了,现在进行更新
		global.DB.Model(this.log).Updates(map[string]any{"content": content}) //像这样按需要去更新就行了 // 已经有日志了,更新一下就行了,避免重复保存
		this.itemList = []string{}                                            //清空日志项列表,保存逻辑上的自洽,不然会多次触发更新
		return
	}

	var itemList []string

	if this.showRequestHeader { //展示请求头,如果是则把请求头加入到日志项列表中
		byteData, _ := json.Marshal(this.c.Request.Header)
		itemList = append(itemList, fmt.Sprintf("<div class=\"log_request_header\"><pre class=\"log_json_body\">%s</pre></div>", string(byteData))) //请求头
	}

	if this.showRequest {
		itemList = append(itemList, fmt.Sprintf("<div class=\"log_request\"><div class=\"log_request_head\"><span class=\"log_request_method delete\">%s</span><span class=\"log_request_path\">%s</span></div><div class=\"log_request_body\"><pre class=\"log_json_body\">%s</pre></div></div>",
			this.c.Request.Method,
			strings.ToLower(this.c.Request.URL.String()), //要做ToLower处理,因为路径是大小写不敏感的
			string(this.requestBody),
		)) //方法,路径,请求体
	}

	//中间出现的一些content,比如日志项等,也要展示出来,以保持日志信息的完整性,避免因为日志项不展示导致日志信息缺失,无法还原当时的操作场景,所以要把日志项插入到请求体和响应体之间,以此来保持日志信息的完整性
	itemList = append(itemList, this.itemList...)

	if this.isMiddlewareSave { //如果是中间件保存,则展示响应头和响应体,否则不展示
		if this.showResponseHeader {
			byteData, _ := json.Marshal(this.ResponseHeader)
			itemList = append(itemList, fmt.Sprintf("<div class=\"log_response_header\"><pre class=\"log_json_body\">%s</pre></div>", string(byteData))) //响应头
		}

		if this.showResponse {
			itemList = append(itemList, fmt.Sprintf("<div class=\"log_response\"><pre class=\"log_json_body\">%s</pre></div>", string(this.responseBody)))
		}

	}

	ipAdd := this.c.ClientIP()
	addr := ip.GetIpAddr(ipAdd)

	var userID uint = 0
	claims, err := jwts.ParseTokenByGin(this.c)
	if err == nil && claims != nil {
		userID = claims.UserID
	}

	log := models.LogModel{
		LogType:     enum.ActionLogType,
		Title:       this.title,
		Content:     strings.Join(itemList, "\n"),
		Level:       this.level,
		UserID:      uint(userID),
		IP:          ipAdd,
		Addr:        addr,
		ServiceName: "操作日志", //TODO:进行定制化的记录
	}

	err = global.DB.Create(&log).Error
	if err != nil {
		logrus.Errorf("保存操作日志失败: %v", err)
		return
	}
	this.log = &log
	this.itemList = []string{} //清空日志项列表,保存逻辑上的自洽,不然会多次触发上面的更新
	return log.ID
}

// NewActionLog 创建新并返回一个保留context状态的操作日志对象
func NewActionLog(c *gin.Context) *ActionLog {
	return &ActionLog{c: c}
}

func GetLog(c *gin.Context) *ActionLog { //通过拿取存储于gin.Context中的日志对象,如果没有就创建一个新的,以此来实现日志对象的复用,保留日志对象的状态,比如请求体和响应体等,避免重复创建日志对象导致日志信息丢失.这是打通日志流的关键
	_log, ok := c.Get("log")
	if !ok {
		return NewActionLog(c)
	}
	log, ok := _log.(*ActionLog)
	if !ok {
		return NewActionLog(c)
	}
	c.Set("saveLog", true)
	return log
}
