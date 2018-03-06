//提供一个分等级的日志系统，建议直接使用全局的对象，而不是另外New一个
package logger

import (
	"fmt"
	"jscfg"
	"log"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	//"sync/atomic"
	"time"
	"timer"
)

const (
	CfgBaseDir  = "../cfg/"
	LogBaseDir  = "../log/"
	LogBaseDir2 = "../logbyday/"
)

//配置表
type stCfg struct {
	LogNumPreFile uint32 //最大条数
	LogLevel      int    //日志级别
}

var cfg stCfg

//基础文件夹（可执行程序目录/../log/应用程序名/）
var sBasePath = ""
var sBasePath2 = ""

//log索引
var iLogIndex uint64 = 0
var lCreateFile sync.Mutex

var iLogIndex2 uint64 = 0
var lCreateFile2 sync.Mutex

//记录到文件
func createLoggerFile(index uint64) {
	lCreateFile.Lock()
	if index < iLogIndex {
		lCreateFile.Unlock()
		return
	}
	iLogIndex = index + 1
	lCreateFile.Unlock()

	//文件夹被删除了？
	if err := os.MkdirAll(sBasePath, os.ModePerm); err != nil {
		return
	}

	stime := time.Now().Format("20060102150405")
	i := 0
	sname := ""
	for {
		sname = sBasePath + stime + "-" + strconv.Itoa(i) + ".log"
		f, err := os.OpenFile(sname, os.O_CREATE|os.O_EXCL|os.O_RDWR, os.ModePerm)
		if err != nil {
			i++
			continue
		}

		loggerTemp := New(f, "", log.LstdFlags|log.Lshortfile, cfg.LogLevel, iLogIndex)
		if globalLogger != nil {
			globalLogger.close()
		}
		globalLogger = loggerTemp

		//标准输出重定向
		os.Stdout = f
		os.Stderr = f

		break
	}

	//定时换
	timeNow := time.Now()
	timeNext := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), timeNow.Hour()+1, 0, 0, 0, time.Local)
	tm := timer.NewTimer(time.Second * time.Duration(timeNext.Unix()-timeNow.Unix()))
	tm.Start(func() {
		tm.Stop()
		createLoggerFile(iLogIndex)
	})
}

//定时读取配置表
func keepLoadCfg() error {
	spath, err := os.Getwd()
	if err != nil {
		return err
	}

	var cfgTemp stCfg
	//读取配置表
	if err := jscfg.ReadJson(path.Join(spath, CfgBaseDir+"logger.json"), &cfgTemp); err != nil {
		return err
	}

	if cfg.LogLevel != cfgTemp.LogLevel && globalLogger != nil {
		globalLogger.SetLevel(cfgTemp.LogLevel)
	}
	cfg = cfgTemp

	return nil
}

func init() {
	_, sfile := path.Split(os.Args[0])

	spath, err := os.Getwd()
	if err != nil {
		panic(err)
		return
	}

	//读取配置表
	if err := jscfg.ReadJson(path.Join(spath, CfgBaseDir+"logger.json"), &cfg); err != nil {
		panic(err)
		return
	}
	sBasePath = path.Join(spath, LogBaseDir, sfile) + "/"
	sBasePath2 = path.Join(spath, LogBaseDir2, sfile) + "/"

	//定时读取配置表，开关日志及数量
	tm := timer.NewTimer(time.Second * 5)
	tm.Start(func() {
		keepLoadCfg()
	})

	createLoggerFile(iLogIndex)
	createLoggerFile2(iLogIndex)
}

var globalLogger *Logger

const (
	DEBUG = iota
	INFO
	WARNING
	ERROR
	FATAL
	NONE
)

var levelNames = []string{
	"DEBUG",
	"INFO",
	"WARNING",
	"ERROR",
	"FATAL",
	"NONE",
}

var levelPrefixes []string

func init() {
	levelPrefixes = make([]string, len(levelNames))
	for i, name := range levelNames {
		levelPrefixes[i] = name + ": "
	}
}

func Debug(format string, args ...interface{}) {
	globalLogger.Output(DEBUG, format, args...)
}

func Info(format string, args ...interface{}) {
	globalLogger.Output(INFO, format, args...)
}

func Warning(format string, args ...interface{}) {
	globalLogger.Output(WARNING, format, args...)
}

func Error(format string, args ...interface{}) {
	globalLogger.Output(ERROR, format, args...)
	// globalLogger2.Output(ERROR, format, args...)
}

func Fatal(format string, args ...interface{}) {
	globalLogger.Output(FATAL, format, args...)
	debug.PrintStack()
	panic(fmt.Sprintf(format, args...))
}

func SetLogger(logger *Logger) {
	globalLogger = logger
}

type Logger struct {
	file    *os.File
	logger  *log.Logger
	level   int
	index   uint64 //索引，防止多个log同时请求创建新的文件
	numbers uint32 //log数量，超出开新文件
}

func New(f *os.File, prefix string, flag, level int, index uint64) *Logger {
	return &Logger{
		file:    f,
		logger:  log.New(f, prefix, flag),
		level:   level,
		index:   index,
		numbers: uint32(0),
	}
}

func (self *Logger) Debug(format string, args ...interface{}) {
	self.Output(DEBUG, format, args...)
}

func (self *Logger) Info(format string, args ...interface{}) {
	self.Output(INFO, format, args...)
}

func (self *Logger) Warning(format string, args ...interface{}) {
	self.Output(WARNING, format, args...)
}

func (self *Logger) Error(format string, args ...interface{}) {
	self.Output(ERROR, format, args...)
}

func (self *Logger) Fatal(format string, args ...interface{}) {
	self.Output(FATAL, format, args...)
	debug.PrintStack()
	panic(fmt.Sprintf(format, args...))
}

//关闭，只调用一次，给30秒的缓冲，肯定都已经写完了
func (self *Logger) close() {
	go func() {
		time.Sleep(time.Second * 30)
		self.file.Close()
	}()
}

// 如果对象包含需要加密的信息（例如密码），请实现Redactor接口
type Redactor interface {
	// 返回一个去处掉敏感信息的示例
	Redacted() interface{}
}

// Redact 返回跟字符串等长的“＊”。
func Redact(s string) string {
	return strings.Repeat("*", len(s))
}

func (self *Logger) Output(level int, format string, args ...interface{}) {
	if self.level > level {
		return
	}
	redactedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if redactor, ok := arg.(Redactor); ok {
			redactedArgs[i] = redactor.Redacted()
		} else {
			redactedArgs[i] = arg
		}
	}
	self.logger.Output(3, levelPrefixes[level]+fmt.Sprintf(format, redactedArgs...))

	//新文件，换新方式了
	//if atomic.AddUint32(&self.numbers, 1) >= cfg.LogNumPreFile {
	//	go createLoggerFile(self.index)
	//}
}

func (self *Logger) SetFlags(flag int) {
	self.logger.SetFlags(flag)
}

func (self *Logger) SetPrefix(prefix string) {
	self.logger.SetPrefix(prefix)
}

func (self *Logger) SetLevel(level int) {
	self.level = level
}

func LogNameToLogLevel(name string) int {
	s := strings.ToUpper(name)
	for i, level := range levelNames {
		if level == s {
			return i
		}
	}
	panic(fmt.Errorf("no log level: %v", name))
}

//一天的错误日志
var globalLogger2 *Logger

//记录到文件
func createLoggerFile2(index uint64) {
	lCreateFile2.Lock()
	if index < iLogIndex2 {
		lCreateFile2.Unlock()
		return
	}
	iLogIndex2 = index + 1
	lCreateFile2.Unlock()
	//文件夹被删除了？
	if err := os.MkdirAll(sBasePath2, os.ModePerm); err != nil {
		return
	}

	stime := time.Now().Format("20060102150405")
	i := 0
	sname := ""
	for {
		sname = sBasePath2 + stime + "-" + strconv.Itoa(i) + ".log"
		f, err := os.OpenFile(sname, os.O_CREATE|os.O_EXCL|os.O_RDWR, os.ModePerm)
		if err != nil {
			i++
			continue
		}

		loggerTemp := New(f, "", log.LstdFlags|log.Lshortfile, cfg.LogLevel, iLogIndex2)
		if globalLogger2 != nil {
			globalLogger2.close()
		}
		globalLogger2 = loggerTemp

		//标准输出重定向
		os.Stdout = f
		os.Stderr = f

		break
	}

	//定时换
	timeNow := time.Now()
	timeNext := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day()+1, 0, 0, 0, 0, time.Local)
	tm := timer.NewTimer(time.Second * time.Duration(timeNext.Unix()-timeNow.Unix()))
	tm.Start(func() {
		tm.Stop()
		createLoggerFile2(iLogIndex)
	})
}
