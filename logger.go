package logger

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	// DEBUG logger level 0
	DEBUG = iota
	// TRACE logger level 1
	TRACE
	// INFO logger level 2
	INFO
	// WARN logger level 3
	WARN
	// ERROR logger level 4
	ERROR
	// FATAL logger level 5
	FATAL
	// LAYOUT timeformat
	LAYOUT = "2006-01-02 15:04:05.000"
)

var (
	// ensure logFile & logConsole implement logger
	_                logger   = (*logFile)(nil)
	_                logger   = (*logConsole)(nil)
	logs             []logger //sharding instance
	consoleOutPut    bool
	fileOutPut       = true
	fileChan         chan map[*os.File]string
	wg               sync.WaitGroup
	megabyteSize     int64 = 100 * 1024 * 1024
	backupTimeFormat       = "2006-01-02T15-04-05.000"
)

// logger define golbal interface
type logger interface {
	setLevel(level int)
	debug(format string, args ...interface{})
	trace(format string, args ...interface{})
	info(format string, args ...interface{})
	wran(format string, args ...interface{})
	error(format string, args ...interface{})
	fatal(format string, args ...interface{})
	close()
	getLevel() int
}

// SetConsole kubernetes app is default with the options
func SetConsole() { consoleOutPut = true }

// UnSetOutFile kubernetes app close file output
func UnSetOutFile() { fileOutPut = false }

// SetMaxSizeMb default maxsize 100Mb
func SetMaxSizeMb(size int64) { megabyteSize = size * 1024 * 1024 }

// Flush ?
func Flush() {
	wg.Wait()
	for _, log := range logs {
		log.close()
	}
}

func shardingInstance() bool {
	if logs == nil {
		logs = make([]logger, 0, 2)
		return false
	}
	return true
}

func getTimeLayout() string {
	return time.Now().Format(LAYOUT)
}

func _getLevelOut(level int, log logger) (out func(string, ...interface{})) {
	switch level {
	case DEBUG:
		out = log.debug
	case TRACE:
		out = log.trace
	case INFO:
		out = log.info
	case WARN:
		out = log.wran
	case ERROR:
		out = log.error
	case FATAL:
		out = log.fatal
	default:
	}
	return out
}

func getLevelString(level int) (out string) {
	switch level {
	case TRACE:
		out = "TRACE"
	case INFO:
		out = "INFO"
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

// return runtime call stack info merge to format
func getCallerStackInfo() (format string) {
	_pc, _file, _line, ok := runtime.Caller(6)
	if ok {
		_func := runtime.FuncForPC(_pc).Name()
		funcName := strings.Split(path.Base(_func), ".")
		fileName := strings.Split(path.Base(_file), string(os.PathSeparator))
		format = fmt.Sprintf("<%s.%s:%d> =>", fileName[len(fileName)-1], funcName[len(funcName)-1], _line)
	}
	return format
}

// directory not exists created
func directory(dir string) {
	if _, e := os.Stat(dir); os.IsNotExist(e) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			fmt.Fprintf(os.Stdout, "directory create error %s\n", dir)
		}
	}
}

func _write(out logger, file *os.File, level int, format string, args ...interface{}) {
	if out.getLevel() > level {
		return
	}

	switch out.(type) {
	case *logConsole:
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

	case *logFile:
		wg.Add(1) // 需要放在管道输入前,Done在管道输出后
		fileChan <- map[*os.File]string{
			file: fmt.Sprintf(fmt.Sprintf(
				"%s [%s] %s %s",
				getTimeLayout(),
				getLevelString(level),
				getCallerStackInfo(),
				format,
			), args...),
		}

	default:
	}
}

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
		}
	}
}

func _rotateFile(file *os.File) (new *os.File) {
	info, err := file.Stat()
	if err != nil {
		panic(err)
	}
	if info.Size() < megabyteSize {
		return file
	}
	new, err = openNew(file.Name())
	if err != nil {
		panic(err)
	}
	return new
}

func NewLogger(cfg map[string]interface{}) {
	exists := shardingInstance()
	if exists {
		return
	}
	var (
		level       int
		logPth      string
		logFile     string
		logChanSize int
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
		_logPth = "logs/"
	} else {
		p, ok := _logPth.(string)
		if !ok {
			panic("configure path value is not string type")
		}
		logPth = p
	}
	if _file, ok := cfg["file"]; !ok {
		_, _tmpfile, _, ok := runtime.Caller(2)
		if ok {
			_file = strings.Split(path.Base(_tmpfile), string(os.PathSeparator))
		}
	} else {
		f, ok := _file.(string)
		if !ok {
			panic("configure file value is not string type")
		}
		logFile = f
	}
	if _chan, ok := cfg["buffer"]; !ok {
		logChanSize = 10000
	} else {
		c, ok := _chan.(int)
		if !ok {
			panic("configure file value is not string type")
		}
		logChanSize = c
	}

	//console
	logs = append(logs, newLogConsole(level))
	//file
	if fileOutPut {
		fileChan = make(chan map[*os.File]string, logChanSize)
		logs = append(logs, newLogFile(level, logPth, logFile))
		go _asyncWrite()
	}
}

func _out(level int, format string, args ...interface{}) {
	if len(logs) < 1 || logs == nil {
		panic("logger instance not init...")
	}
	for _, log := range logs {
		if _, ok := log.(*logConsole); ok && consoleOutPut {
			(_getLevelOut(level, log))(format, args...)
		}
		if _, ok := log.(*logFile); ok && fileOutPut {
			(_getLevelOut(level, log))(format, args...)
		}
	}
}

// Debug
func Debug(format string, args ...interface{}) { _out(DEBUG, format, args...) }

// Trace
func Trace(format string, args ...interface{}) { _out(TRACE, format, args...) }

// Info ?
func Info(format string, args ...interface{}) { _out(INFO, format, args...) }

// Warn ?
func Warn(format string, args ...interface{}) { _out(WARN, format, args...) }

// Error ?
func Error(format string, args ...interface{}) { _out(ERROR, format, args...) }

// Fatal ?
func Fatal(format string, args ...interface{}) { _out(FATAL, format, args...) }

func openNew(name string) (*os.File, error) {

	mode := os.FileMode(0644)
	info, err := os.Stat(name)

	if err == nil {
		mode = info.Mode()
		newname := backupName(name, true)
		if err := os.Rename(name, newname); err != nil {
			return nil, fmt.Errorf("can't rename log file: %s", err)
		}
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, mode)

	if err != nil {
		return f, fmt.Errorf("can't open new logfile: %s", err)
	}

	return f, err
}

func backupName(name string, local bool) string {
	dir := filepath.Dir(name)
	filename := filepath.Base(name)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	t := time.Now()
	if !local {
		t = t.UTC()
	}
	timestamp := t.Format(backupTimeFormat)
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", prefix, timestamp, ext))
}

// logFile output info to file [info,warn]
type logFile struct {
	level int
	path  string
	name  string
	file  *os.File
	warn  *os.File
	lock  *sync.Mutex
}

// fileNew opens a new log file for writing, moving any old log file out of the
func (lf *logFile) fileNew() (err error) {
	lf.file, err = openNew(fmt.Sprintf("%s%s-info.log", lf.path, lf.name))
	return err
}

func (lf *logFile) warnNew() (err error) {
	lf.warn, err = openNew(fmt.Sprintf("%s%s-warn.log", lf.path, lf.name))
	return err
}

// NewLogFile ?
func newLogFile(level int, path string, file string) logger {
	// directory check if not exists then created
	directory(path)
	logfile := &logFile{
		level: level,
		path:  path,
		name:  file,
		lock:  &sync.Mutex{},
	}

	err := logfile.fileNew()
	if err != nil {
		panic(err)
	}
	err = logfile.warnNew()
	if err != nil {
		panic(err)
	}

	return logfile
}

func (lf *logFile) setLevel(level int) {
	if level < DEBUG || level > FATAL {
		level = DEBUG
	}
	lf.level = level
}

func (lf *logFile) getLevel() int {
	return lf.level
}

func (lf *logFile) debug(format string, args ...interface{}) {
	lf.lock.Lock()
	defer lf.lock.Unlock()
	lf.file = _rotateFile(lf.file)
	_write(lf, lf.file, DEBUG, format, args...)
}

func (lf *logFile) trace(format string, args ...interface{}) {
	lf.lock.Lock()
	defer lf.lock.Unlock()
	lf.file = _rotateFile(lf.file)
	_write(lf, lf.file, TRACE, format, args...)
}

func (lf *logFile) info(format string, args ...interface{}) {
	lf.lock.Lock()
	defer lf.lock.Unlock()
	lf.file = _rotateFile(lf.file)
	_write(lf, lf.file, INFO, format, args...)
}

func (lf *logFile) wran(format string, args ...interface{}) {
	lf.lock.Lock()
	defer lf.lock.Unlock()
	lf.file = _rotateFile(lf.file)
	_write(lf, lf.file, WARN, format, args...)
}

func (lf *logFile) error(format string, args ...interface{}) {
	lf.lock.Lock()
	defer lf.lock.Unlock()
	lf.warn = _rotateFile(lf.warn)
	_write(lf, lf.warn, ERROR, format, args...)
}

func (lf *logFile) fatal(format string, args ...interface{}) {
	lf.lock.Lock()
	defer lf.lock.Unlock()
	lf.warn = _rotateFile(lf.warn)
	_write(lf, lf.warn, FATAL, format, args...)
}

func (lf *logFile) close() {
	lf.lock.Lock()
	defer lf.lock.Unlock()
	switch {
	case lf.file != nil:
		lf.file.Close()
		fallthrough
	case lf.warn != nil:
		lf.warn.Close()
	}
}

// logConsole
type logConsole struct {
	level int
	file  *os.File
	warn  *os.File
}

func newLogConsole(level int) logger {
	return &logConsole{
		level: level,
		file:  os.Stdout,
		warn:  os.Stderr,
	}
}

func (lc *logConsole) setLevel(level int) {
	if level < DEBUG || level > FATAL {
		level = DEBUG
	}
	lc.level = level
}

func (lc *logConsole) debug(format string, args ...interface{}) {
	_write(lc, lc.file, DEBUG, format, args...)
}

func (lc *logConsole) trace(format string, args ...interface{}) {
	_write(lc, lc.file, TRACE, format, args...)
}
func (lc *logConsole) info(format string, args ...interface{}) {
	_write(lc, lc.file, INFO, format, args...)
}

func (lc *logConsole) wran(format string, args ...interface{}) {
	_write(lc, lc.file, WARN, format, args...)
}

func (lc *logConsole) error(format string, args ...interface{}) {
	_write(lc, lc.warn, ERROR, format, args...)
}

func (lc *logConsole) fatal(format string, args ...interface{}) {
	_write(lc, lc.warn, FATAL, format, args...)
}

func (lc *logConsole) close() {
	if lc.file != nil {
		lc.file.Close()
	}
}

func (lc *logConsole) getLevel() int { return lc.level }
