// core/init_logrus.go
package core

import (
	"StarDreamerCyberNook/global"
	"bytes"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
)

// 颜色
const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

// ============ 终端用：带颜色 ============
type ConsoleFormatter struct{}

func (t *ConsoleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format("2006-01-02 15:04:05")

	if entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
		// 带 ANSI 颜色
		fmt.Fprintf(b, "[%s] \x1b[%dm[%s]\x1b[0m %s %s %s\n",
			timestamp, levelColor, entry.Level, fileVal, funcVal, entry.Message)
	} else {
		fmt.Fprintf(b, "[%s] \x1b[%dm[%s]\x1b[0m %s\n",
			timestamp, levelColor, entry.Level, entry.Message)
	}
	return b.Bytes(), nil
}

// ============ 文件用：纯文本，无颜色 ============
type FileFormatter struct{}

func (t *FileFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format("2006-01-02 15:04:05")

	if entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
		// 纯文本，无颜色代码
		fmt.Fprintf(b, "[%s] [%s] %s %s %s\n",
			timestamp, entry.Level, fileVal, funcVal, entry.Message)
	} else {
		fmt.Fprintf(b, "[%s] [%s] %s\n",
			timestamp, entry.Level, entry.Message)
	}
	return b.Bytes(), nil
}

// ============ 文件 Hook（关键修改） ============
type FileDateHook struct {
	file     *os.File
	logPath  string
	fileDate string
	appName  string
	// 新增：文件专用 formatter
	formatter *FileFormatter
}

func (hook FileDateHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook FileDateHook) Fire(entry *logrus.Entry) error {
	timer := entry.Time.Format("2006-01-02")

	// 使用 FileFormatter 格式化（无颜色）
	line, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}

	if hook.fileDate == timer {
		hook.file.Write(line)
		return nil
	}

	// 日期切换，创建新文件
	hook.file.Close()
	os.MkdirAll(fmt.Sprintf("%s/%s", hook.logPath, timer), os.ModePerm)
	filename := fmt.Sprintf("%s/%s/%s.log", hook.logPath, timer, hook.appName)

	hook.file, _ = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	hook.fileDate = timer
	hook.file.Write(line)
	return nil
}

func InitFile(logPath, appName string) {
	fileDate := time.Now().Format("2006-01-02")
	err := os.MkdirAll(fmt.Sprintf("%s/%s", logPath, fileDate), os.ModePerm)
	if err != nil {
		logrus.Error(err)
		return
	}

	filename := fmt.Sprintf("%s/%s/%s.log", logPath, fileDate, appName)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		logrus.Error(err)
		return
	}

	// 传入 FileFormatter
	fileHook := FileDateHook{
		file:      file,
		logPath:   logPath,
		fileDate:  fileDate,
		appName:   appName,
		formatter: &FileFormatter{},
	}
	logrus.AddHook(&fileHook)
}

func InitLogrus() {
	logrus.SetOutput(os.Stdout)              // 设置输出类型
	logrus.SetReportCaller(true)             // 开启返回函数名和行号
	logrus.SetFormatter(&ConsoleFormatter{}) // 终端用带颜色的 Formatter
	logrus.SetLevel(logrus.DebugLevel)       // 设置最低的 Level
	l := global.Config.Log
	InitFile(l.Dir, l.App)
}
