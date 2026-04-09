// service/log_service/runtime_log.go
// 运行时日志
package log_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type runtimeLog struct { //运行时日志.和操作日志不同的是,该日志实例化之后将长时间使用
	level           enum.LogLevel   //日志级别
	title           string          //日志标题
	itemList        []string        //日志内容项
	serviceName     string          //服务名称
	runtimeDateType RuntimeDateType //按照时间长度分割
}

type RuntimeDateType int //按照时间长度分割
const (
	RuntimeDateHour  RuntimeDateType = 1
	RuntimeDateDay   RuntimeDateType = 2
	RuntimeDateWeek  RuntimeDateType = 3
	RuntimeDateMonth RuntimeDateType = 4
)

func (this RuntimeDateType) String() string { //原函数名GetSqlTime
	switch this {
	case RuntimeDateHour:
		return "interval 1 HOUR"
	case RuntimeDateDay:
		return "interval 1 DAY"
	case RuntimeDateWeek:
		return "interval 1 WEEK"
	case RuntimeDateMonth:
		return "interval 1 MONTH"
	default:
		return "interval 1 DAY" //默认按天分割
	}
}

func (this *runtimeLog) Save() {

	var log models.LogModel
	this.SetNowTime()

	global.DB.Find(&log, fmt.Sprintf("service_name=? and log_type=? and created_at >= date_sub(now(),%s)",
		this.runtimeDateType.String()), this.serviceName, enum.RuntimeLogType)
	if log.ID != 0 {
		//说明有数据,需要更新
		c := strings.Join(this.itemList, "\n")
		newContent := log.Content + "\n" + c
		global.DB.Model(&log).Updates(map[string]any{"content": newContent})
		this.itemList = []string{}
	}
	//判断是创建还是更新
	log2 := &models.LogModel{
		LogType:     enum.RuntimeLogType,
		Level:       this.level,
		Title:       this.title,
		Content:     strings.Join(this.itemList, "\n"), //这里是运行时日志的内容
		ServiceName: this.serviceName,
	}
	err := global.DB.Create(log2).Error //写入数据库
	if err != nil {
		logrus.Errorf("创建日志失败: %s", err)
		return
	}
}

func (this *runtimeLog) NewRuntimeLog(serviceName string, runtimeDateType RuntimeDateType) *runtimeLog {
	return &runtimeLog{
		serviceName:     serviceName,
		runtimeDateType: runtimeDateType,
	}
}
func (this *runtimeLog) setItem(lable string, value any, LogLevelType enum.LogLevel) { //设置项,是各等级设置项的具体实现
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
func (this *runtimeLog) SetItem(lable string, value any) { //默认等级为info
	this.setItem(lable, value, enum.LogInfoLevel)
}
func (this *runtimeLog) SetItemInfo(lable string, value any) { //设置并加入info等级的项
	this.setItem(lable, value, enum.LogInfoLevel)
}
func (this *runtimeLog) SetItemWarn(lable string, value any) { //设置warn等级的项
	this.setItem(lable, value, enum.LogWarnLevel)
}
func (this *runtimeLog) SetItemError(lable string, value any) { //设置error等级的项
	this.setItem(lable, value, enum.LogErrorLevel)
}

func (this *runtimeLog) SetError(label string, err error) { //设置error等级的项,并追踪错误栈
	var msg = make([]byte, 1024)
	n := runtime.Stack(msg, false)
	logrus.Errorf("出错:%s", err.Error())
	this.itemList = append(this.itemList, fmt.Sprintf("<div class=\"log_error\"><div class=\"line\"><div class=\"label\">%s</div><div class=\"value\">%s</div><div class=\"type\">%t</div></div><div class=\"stack\">%+v</div></div>",
		label, err, err, msg[:n]))
}

func (this *runtimeLog) SetTitle(title string) { //设置标题
	this.title = title
}

func (this *runtimeLog) SetLink(label string, href string) { //设置链接项
	this.itemList = append(this.itemList, fmt.Sprintf("<div class=\"log_item Link\"><div class=\"log_item_label\">%s</div><div class=\"Log_item_content\"><a href=\"%s\" target=\"blank\">%s</a></div></div>", label, href, href))
}

func (this *runtimeLog) SetImage(src string) { //设置图片项
	this.itemList = append(this.itemList, fmt.Sprintf("<div class=\"log_Image\"><img src=\"%s\" alt=\"\"/></div>", src))
}

func (this *runtimeLog) SetLevel(level enum.LogLevel) { //设置日志项等级
	this.level = level
}

func (this *runtimeLog) SetNowTime() {
	this.itemList = append(this.itemList, fmt.Sprintf("<div class=\"log_time\">%s</div>", time.Now().Format("2006-01-02 15:04:05")))
}
