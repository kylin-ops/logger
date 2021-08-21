package logger

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
	"time"
)

func caller(skip int) (string, bool) {
	pc, file, line, ok := runtime.Caller(skip)
	pcName := runtime.FuncForPC(pc).Name()
	msg := fmt.Sprintf("%s %d %s\n", file, line, pcName)
	return msg, ok
}

func fileTrack() string {
	var msg string
	for i := 0; i < 32; i++ {
		m, ok := caller(i)
		if ok == false {
			break
		}
		if strings.Index(m, "rifflock/lfshook") == -1&strings.Index(m, "sirupsen/logrus") {
			msg = msg + m
		}
	}
	return msg
}

// 自定义日志格式
type myFormatter struct{}

func (s myFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.999")
	level := fmt.Sprintf("%-7s", entry.Level.String())
	message := entry.Message
	//var msg string
	//if level == "ERROR" {
	//	file := fileTrack()
	//	// file := debug.Stack()
	//	msg = fmt.Sprintf("[%s]  [%-6s]  %s\n %s", timestamp, level, message, file)
	//} else {
	//	msg = fmt.Sprintf("[%s]  [%-6s]  %s\n", timestamp, level, message)
	//}
	msg := fmt.Sprintf("[%-23s]  [%-6s]  %s\n", timestamp, level, message)
	return []byte(msg), nil
}

func setLogLevel(level string) logrus.Level {
	level = strings.ToLower(level)
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warm":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "trace":
		return logrus.TraceLevel
	case "fatal":
		return logrus.FatalLevel
	default:
		logrus.Warn("日志级别设置错误，使用默认日志级别:\"info\"")
		return logrus.InfoLevel
	}
}

func NewLogger(level, logPath string, rollTime time.Duration, logCount int, isConsole bool) (*logrus.Logger, error) {
	var err error
	var log = logrus.New()
	log.SetReportCaller(true)
	// 设置日志级别为xx以及以上
	log.SetLevel(setLogLevel(level))
	//log.AddHook(&defaultFieldHook{})
	// 设置日志格式为json格式
	// log.SetFormatter(&logrus.JSONFormatter{
	// 	// PrettyPrint: true,//格式化json
	// 	TimestampFormat: "2006-01-02 15:04:05",//时间格式化
	// })
	//log.SetFormatter(&logrus.TextFormatter{
	//	ForceColors:               true,
	//	EnvironmentOverrideColors: true,
	//	// FullTimestamp:true,
	//	TimestampFormat: "2006-01-02 15:04:05", //时间格式化
	//	// DisableLevelTruncation:true,
	//})
	log.SetFormatter(myFormatter{})
	// 设置将日志输出到标准输出（默认的输出为stdout，标准错误）
	// 日志消息输出可以是任意的io.writer类型
	// file, _ := os.OpenFile("/tmp/info.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if isConsole {
		log.SetOutput(os.Stdout)
	} else {
		file, err := os.OpenFile("/dev/null", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		log.SetOutput(file)
	}
	if logPath != "" {
		writer, err := rotatelogs.New(
			//这是分割代码的命名规则，要和下面WithRotationTime时间精度一致
			logPath+".%Y%m%d%H%M%S",
			// WithLinkName为最新的日志建立软连接，以方便随着找到当前日志文件。
			rotatelogs.WithLinkName(logPath),
			//文件切割之间的间隔。默认情况下，日志每86400秒/一天旋转一次。注意:记住要利用时间。持续时间值。
			rotatelogs.WithRotationTime(rollTime),
			// WithMaxAge和WithRotationCount二者只能设置一个，
			// WithMaxAge设置文件清理前的最长保存时间，
			// WithRotationCount设置文件清理前最多保存的个数。 默认情况下，此选项是禁用的。
			// rotatelogs.WithMaxAge(time.Second*30), //默认每7天清除下日志文件
			rotatelogs.WithRotationCount(uint(logCount)),
			//rotatelogs.WithMaxAge(-1),       //需要手动禁用禁用  默认情况下不清除日志，
			// rotatelogs.WithRotationCount(2), //清除除最新2个文件之外的日志，默认禁用
		)
		if err != nil {
			return nil, err
		}

		lfsHook := lfshook.NewHook(lfshook.WriterMap{
			logrus.TraceLevel: writer,
			logrus.DebugLevel: writer,
			logrus.InfoLevel:  writer,
			logrus.WarnLevel:  writer,
			logrus.ErrorLevel: writer,
			logrus.FatalLevel: writer,
			logrus.PanicLevel: writer,
		}, &myFormatter{})
		log.AddHook(lfsHook)
	}
	return log, err
}
