package logger

import "testing"
import "os"
import . "github.com/smartystreets/goconvey/convey"

func TestFileLogger(t *testing.T) {
	Convey("first test", t, func() {
		return
	})
}

func TestDirectoryCheck(t *testing.T) {
	Convey("create test directory in ./log2", t, func() {

		directory("./log2")

		Convey("check exists?", func() {

			if _, e := os.Stat("./log2"); os.IsNotExist(e) {
				So(e, ShouldBeError)
			} else {
				os.Remove("./log2")
			}
		})

	})
}
