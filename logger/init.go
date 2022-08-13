package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/env"
	"github.com/rickslab/ares/util"
	"github.com/sirupsen/logrus"
)

func init() {
	if env.IsDebug() {
		return
	}

	i := strings.LastIndex(os.Args[0], "/")
	name := os.Args[0][i+1:]

	traceLogHook, err := NewFileRotateHook(fmt.Sprintf("%s/trace.log", name), logrus.WarnLevel, logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel)
	util.AssertError(err)
	logrus.AddHook(traceLogHook)

	errorLogHook, err := NewFileRotateHook(fmt.Sprintf("%s/error.log", name), logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel)
	util.AssertError(err)
	logrus.AddHook(errorLogHook)

	graylogAddress := config.YamlEnv().GetString("service.graylog")
	if graylogAddress != "" {
		gelfHook, err := NewUdpGelfHook(graylogAddress, logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel)
		util.AssertError(err)
		logrus.AddHook(gelfHook)
	}

	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetFormatter(&formatter{})
}
