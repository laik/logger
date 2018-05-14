package main

import logger "github.com/laik/logger"

/*
	后续实现功能:
	[√]	1.要初始化2个实例来调用改成一个方法(标准输出窗口控制开关) 日志库的封装
	[]	2.目录不存在问题
	[]	3.日志分割
	[√]	4.单例模式(防止在多个层级初始化)
	[√]	5.异步写日志
	[]  6.切分方式(小时/天/星期/月/体积)
*/

func output() {
	// _c_log := logger.NewLogConsole(logger.DEBUG)
	// _f_log := logger.NewLogFile(logger.DEBUG, "logs/", "logger-demo")

	// _c_log.Debug("console output %s\n", "test console logger out...")
	// _f_log.Debug("console output %s\n", "test file logger out...")

	cfg := make(map[string]interface{}, 0)

	cfg["path"] = "logs/"
	cfg["file"] = "test"
	cfg["level"] = logger.DEBUG
	cfg["buffer"] = 100000

	logger.InitLogger(cfg)

	logger.SetConsole()

	logger.Debug("test debug log out %s\n", "test1")
	logger.Trace("test trace log out %s\n", "test1")
	logger.Info("test info log out %s\n", "test1")
	logger.Warn("test warn log out %s\n", "test1")
	logger.Error("test error log out %s\n", "test1")
	logger.Fatal("test fatal log out %s\n", "test1")

	logger.UnsetConsole()

	logger.Debug("test debug log out %s\n", "test2")
	logger.Trace("test trace log out %s\n", "test2")
	logger.Info("test info log out %s\n", "test2")
	logger.Warn("test warn log out %s\n", "test2")
	logger.Error("test error log out %s\n", "test2")
	logger.Fatal("test fatal log out %s\n", "test2")

}

func main() {
	defer logger.Flush()
	output()
}
