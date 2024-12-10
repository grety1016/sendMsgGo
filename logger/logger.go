package logger

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

//#region 日志记录包

// DB 日志初始化
var logOnce sync.Once

type LogFormatter struct{}

var re = regexp.MustCompile(`\{\{(.*)\}\}`)

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b bytes.Buffer

	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	level := fmt.Sprintf("%-7s", entry.Level.String())

	logLine := fmt.Sprintf("[%s][%s]%s", timestamp, level, entry.Message)

	keys := make([]string, 0, len(entry.Data))
	for key := range entry.Data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	if len(entry.Data) > 0 {
		logLine += " {"
		for i, key := range keys {
			if i > 0 {
				logLine += ","
			}
			logLine += fmt.Sprintf(" %s=%v", key, entry.Data[key])
		}
		logLine += " }"
	}

	logLine += "\n"
	b.WriteString(logLine)

	return b.Bytes(), nil
}

func ToLogFields(data interface{}) logrus.Fields {
	fields := logrus.Fields{}
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	for i := 0; i < v.NumField(); i++ {
		fields[t.Field(i).Name] = v.Field(i).Interface()
	}
	return fields
}

func Init() {
	logOnce.Do(func() {
		currentDate := time.Now().Format("2006-01-02")
		logFile := &lumberjack.Logger{
			Filename:   fmt.Sprintf("./LOGFILES/DBLog_%s.log", currentDate),
			MaxSize:    100,
			MaxBackups: 0,
			MaxAge:     0,
			Compress:   true,
		}

		logrus.SetOutput(logFile)
		logrus.SetFormatter(&LogFormatter{})
		logrus.SetLevel(logrus.InfoLevel)
	})
}

// 格式化 sql.Named 参数的函数，并去掉多余的 {}
func FormatNamedArgs(args ...interface{}) string {
	if len(args) == 0 || (len(args) == 1 && args[0] == nil) {
		return "nil"
	}
	var formattedArgs []string

	processArg := func(arg interface{}) {
		argStr := fmt.Sprintf("%v", arg)
		if namedArg, ok := arg.(sql.NamedArg); ok {
			argStr = fmt.Sprintf("{%s:%v}", namedArg.Name, namedArg.Value)
		}
		argStr = re.ReplaceAllString(argStr, `{$1}`)
		formattedArgs = append(formattedArgs, argStr)
	}

	if len(args) == 1 && reflect.TypeOf(args[0]).Kind() == reflect.Slice {
		slice := reflect.ValueOf(args[0])
		for i := 0; i < slice.Len(); i++ {
			processArg(slice.Index(i).Interface())
		}
	} else {
		for _, arg := range args {
			processArg(arg)
		}
	}

	return strings.Join(formattedArgs, " ")
}

// HTTP 日志初始化
var httpLogOnce sync.Once
var httpLogger *logrus.Logger

type HTTPLogFormatter struct{}

func (f *HTTPLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	level := fmt.Sprintf("%-7s", entry.Level.String())
	logLine := fmt.Sprintf("[%s][%s]%s\n", timestamp, level, entry.Message)
	return []byte(logLine), nil
}

func InitHTTPLogger() *logrus.Logger {
	httpLogOnce.Do(func() {
		httpLogger = logrus.New()
		currentDate := time.Now().Format("2006-01-02")
		logFile := &lumberjack.Logger{
			Filename:   fmt.Sprintf("./LOGFILES/HTTPLog_%s.log", currentDate),
			MaxSize:    100,
			MaxBackups: 0,
			MaxAge:     0,
			Compress:   true,
		}
		httpLogger.SetOutput(logFile)
		httpLogger.SetFormatter(&HTTPLogFormatter{})
		httpLogger.SetLevel(logrus.InfoLevel)
	})
	return httpLogger
}

//#endregion 日志记录包
