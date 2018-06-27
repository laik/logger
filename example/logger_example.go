package main

import (
	logger "github.com/laik/logger"
)

/*
	后续实现功能:
	[√]	1.要初始化2个实例来调用改成一个方法(标准输出窗口控制开关) 日志库的封装
	[√]	2.目录不存在问题
	[√]	3.日志分割 // 当前以mb分割
	[√]	4.单例模式(防止在多个层级初始化)
	[√]	5.异步写日志
	[X]  6.切分方式(小时/天/星期/月/体积) //discard
	[√] 7.启动备份上一次的日志文件
*/

func output() {

	cfg := make(map[string]interface{}, 0)

	cfg["path"] = "logs/"
	cfg["file"] = "test"
	cfg["level"] = logger.DEBUG
	cfg["buffer"] = 100000

	logger.NewLogger(cfg)
	logger.SetMaxSizeMb(3)
	logger.SetConsole()

	logger.Debug("test debug log out %s\n", "test1")
	logger.Trace("test trace log out %s\n", "test1")
	logger.Info("test info log out %s\n", "test1")
	logger.Warn("test warn log out %s\n", "test1")
	logger.Error("test error log out %s\n", "test1")
	logger.Fatal("test fatal log out %s\n", "test1")

	logger.UnsetConsole()

	for i := 0; i < 1000; i++ {
		logger.Debug("test debug log out %s\n", "test2")
		logger.Trace("test trace log out %s\n", "test2")
		logger.Info("test info log out %s\n", "test2")
		logger.Warn("test warn log out %s\n", "test2")
		logger.Error("test error log out %s\n", "test2")
		logger.Fatal("test fatal log out %s\n", "test2")
	}

	/*
		output:
			1.0M    logs/test-info-2018-06-11T15-23-22.512.log
			1.0M    logs/test-info-2018-06-11T15-23-22.642.log
			792K    logs/test-info.log
			1.0M    logs/test-warn-2018-06-11T15-23-22.639.log
			420K    logs/test-warn.log
	*/

}

func main() {
	defer logger.Flush()
	output()
}
