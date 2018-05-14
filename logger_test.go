package logger

import "testing"

func TestFileLogger(t *testing.T) {

	_logger := NewLogFile(DEBUG, "example/logs/", "test")

	defer _logger.Close()

	_logger.Debug("deubg %s\n", "aaa")

	_logger.SetLevel(ERROR)

	_logger.Error("error %s\n", "aaa")

	_logger.SetLevel(TRACE)

	_logger.Trace("trace %s\n", "aaa")

	_c_logger := NewLogConsole(DEBUG)

	_c_logger.SetLevel(DEBUG)

	_c_logger.Debug("c_debug %s\n", "dsa")

	_c_logger.SetLevel(ERROR)

	_c_logger.Error("c_error %s\n", "dsa")

}
