// TODO:日志类型和日志级别可以根据实际需求进行调整和扩展,后期调整一下
package enum

type LogType int8

const (
	LoginLogType   LogType = 1 //登录日志
	ActionLogType  LogType = 2 //操作日志
	RuntimeLogType LogType = 3 //运行时日志
)

type LogLevel int8

const (
	LogInfoLevel  LogLevel = 1 //信息日志
	LogWarnLevel  LogLevel = 2 //警告日志
	LogErrorLevel LogLevel = 3 //错误日志
)

func (this LogType) String() string {
	switch this {
	case LoginLogType:
		return "LoginLog"
	case ActionLogType:
		return "ActionLog"
	case RuntimeLogType:
		return "RuntimeLog"
	default:
		return "UnknownLogType"
	}
}

func (this LogLevel) String() string {
	switch this {
	case LogInfoLevel:
		return "Info"
	case LogWarnLevel:
		return "Warn"
	case LogErrorLevel:
		return "Error"
	default:
		return "UnknownLogLevel"
	}
}
