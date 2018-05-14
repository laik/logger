package logger

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	DEBUG = iota
	TRACE
	INFO
	WARN
	ERROR
	FATAL
	LAYOUT = "2006-01-02 15:04:05.999"
)

var (
	logs          []Logger //sharding instance
	consoleOutPut bool
	fileChan      chan map[*os.File]string
	wg            sync.WaitGroup
)

func SetConsole() { consoleOutPut = true }

func UnsetConsole() { consoleOutPut = false }

func shardingInstance() bool {
	if logs == nil {
		logs = make([]Logger, 0, 2)
		return false
	}
	return true
}

func InitLogger(cfg map[string]interface{}) {

	exists := shardingInstance()

	if exists {
		return
	}

	var (
		level   int
		logPth  string
		logFile string
		logChan int
	)

	if _level, ok := cfg["level"]; !ok {
		level = DEBUG
	} else {
		v, ok := _level.(int)
		if !ok {
			panic("configure level value is not int type")
		}
		level = v
	}

	if _logPth, ok := cfg["path"]; !ok {
		panic("configure not define log file path")
	} else {
		p, ok := _logPth.(string)
		if !ok {
			panic("configure path value is not string type")
		}
		logPth = p
	}

	if _file, ok := cfg["file"]; !ok {
		panic("configure not define log file value")
	} else {
		f, ok := _file.(string)
		if !ok {
			panic("configure file value is not string type")
		}
		logFile = f
	}

	if _chan, ok := cfg["buffer"]; !ok {
		logChan = 10000
	} else {
		c, ok := _chan.(int)
		if !ok {
			panic("configure file value is not string type")
		}
		logChan = c
	}

	fileChan = make(chan map[*os.File]string, logChan)

	logs = append(logs, NewLogConsole(level))

	logs = append(logs, NewLogFile(level, logPth, logFile))

	go _asyncWrite()

}

func _getLevelOut(level int, log Logger) (out func(string, ...interface{})) {
	switch level {
	case DEBUG:
		out = log.Debug
	case TRACE:
		out = log.Trace
	case INFO:
		out = log.Info
	case WARN:
		out = log.Wran
	case ERROR:
		out = log.Error
	case FATAL:
		out = log.Fatal
	default:
	}
	return out
}

func _out(level int, format string, args ...interface{}) {

	if len(logs) < 1 {
		panic("logger instance not init...")
	}

	for _, log := range logs {
		if _, ok := log.(*LogConsole); ok && consoleOutPut {
			(_getLevelOut(level, log))(format, args...)
		}
		if _, ok := log.(*LogFile); ok {
			(_getLevelOut(level, log))(format, args...)
		}
	}

}

func Debug(format string, args ...interface{}) { _out(DEBUG, format, args...) }

func Trace(format string, args ...interface{}) { _out(TRACE, format, args...) }

func Info(format string, args ...interface{}) { _out(INFO, format, args...) }

func Warn(format string, args ...interface{}) { _out(WARN, format, args...) }

func Error(format string, args ...interface{}) { _out(ERROR, format, args...) }

func Fatal(format string, args ...interface{}) { _out(FATAL, format, args...) }

func getLevelString(level int) (out string) {
	switch level {
	case TRACE:
		out = "TRACE"
	case WARN:
		out = "WARN"
	case ERROR:
		out = "ERROR"
	case FATAL:
		out = "FATAL"
	default:
		out = "DEBUG"
	}
	return out
}

func getCallerStackInfo() (format string) {
	_pc, _file, _line, ok := runtime.Caller(6)
	if ok {

		_func := runtime.FuncForPC(_pc).Name()

		_func_name := strings.Split(path.Base(_func), ".")

		_file_name := strings.Split(path.Base(_file), string(os.PathSeparator))

		format = fmt.Sprintf("<%s.%s:%d> =>", _file_name[len(_file_name)-1], _func_name[len(_func_name)-1], _line)
	}
	return format
}

func getTimeLayout() string { return time.Now().Format(LAYOUT) }

func Flush() { wg.Wait() }

func _asyncWrite() {
	for {
		select {
		case _msg, ok := <-fileChan:
			if ok {
				for file, value := range _msg {
					file.WriteString(value)
				}
				wg.Done()
			}
		default:
		}
	}
}

func _write(out Logger, file *os.File, level int, format string, args ...interface{}) {

	if out.GetLevel() > level {
		return
	}

	_, ok := out.(*LogConsole)

	if ok {
		fmt.Fprintf(
			file,
			fmt.Sprintf(
				"%s [%s] %s %s",
				getTimeLayout(),
				getLevelString(level),
				getCallerStackInfo(),
				format,
			),
			args...,
		)
		return
	}

	_, ok2 := out.(*LogFile)

	if ok2 {
		_format := fmt.Sprintf(
			"%s [%s] %s %s",
			getTimeLayout(),
			getLevelString(level),
			getCallerStackInfo(),
			format,
		)
		fileChan <- map[*os.File]string{
			file: fmt.Sprintf(_format, args...),
		}
		wg.Add(1)
	}

}

type Logger interface {
	SetLevel(level int)
	Debug(format string, args ...interface{})
	Trace(format string, args ...interface{})
	Info(format string, args ...interface{})
	Wran(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	Close()
	GetLevel() int
}

type LogFile struct {
	level int
	path  string
	name  string
	file  *os.File
	warn  *os.File
}

func NewLogFile(level int, path string, file string) Logger {

	logfile, err := os.OpenFile(fmt.Sprintf("%s%s-info.log", path, file), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)

	if err != nil {
		panic(fmt.Sprintf("open file %s error failed on logger init, error %v\n", file, err))
	}

	warnfile, err := os.OpenFile(fmt.Sprintf("%s%s-warn.log", path, file), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(fmt.Sprintf("open file %s error failed on logger init, error %v\n", file, err))
	}

	return &LogFile{
		level: level,
		path:  path,
		name:  file,
		file:  logfile,
		warn:  warnfile,
	}
}

func (lf *LogFile) SetLevel(level int) {
	if level < DEBUG || level > FATAL {
		level = DEBUG
	}
	lf.level = level
}

func (lf *LogFile) GetLevel() int {
	return lf.level
}

func (lf *LogFile) Debug(format string, args ...interface{}) {
	_write(lf, lf.file, DEBUG, format, args...)
}

func (lf *LogFile) Trace(format string, args ...interface{}) {
	_write(lf, lf.file, TRACE, format, args...)
}

func (lf *LogFile) Info(format string, args ...interface{}) {
	_write(lf, lf.file, INFO, format, args...)
}

func (lf *LogFile) Wran(format string, args ...interface{}) {
	_write(lf, lf.file, WARN, format, args...)
}

func (lf *LogFile) Error(format string, args ...interface{}) {
	_write(lf, lf.warn, ERROR, format, args...)
}

func (lf *LogFile) Fatal(format string, args ...interface{}) {
	_write(lf, lf.warn, FATAL, format, args...)
}

func (lf *LogFile) Close() { lf.file.Close(); lf.warn.Close() }

type LogConsole struct {
	level int
	file  *os.File
	warn  *os.File
}

func NewLogConsole(level int) Logger {
	return &LogConsole{
		level: level,
		file:  os.Stdout,
		warn:  os.Stderr,
	}
}

func (lc *LogConsole) SetLevel(level int) {
	if level < DEBUG || level > FATAL {
		level = DEBUG
	}
	lc.level = level
}

func (lc *LogConsole) Debug(format string, args ...interface{}) {
	_write(lc, lc.file, DEBUG, format, args...)
}

func (lc *LogConsole) Trace(format string, args ...interface{}) {
	_write(lc, lc.file, TRACE, format, args...)
}
func (lc *LogConsole) Info(format string, args ...interface{}) {
	_write(lc, lc.file, INFO, format, args...)
}

func (lc *LogConsole) Wran(format string, args ...interface{}) {
	_write(lc, lc.file, WARN, format, args...)
}

func (lc *LogConsole) Error(format string, args ...interface{}) {
	_write(lc, lc.warn, ERROR, format, args...)
}

func (lc *LogConsole) Fatal(format string, args ...interface{}) {
	_write(lc, lc.file, FATAL, format, args...)
}

func (lc *LogConsole) Close() {}

func (lc *LogConsole) GetLevel() int { return lc.level }
