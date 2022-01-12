package env

import (
	"os"
	"path/filepath"
)

var (
	debugFlag bool
	envFlag   string
	confPath  string
	logPath   string
)

func init() {
	if os.Getenv("DEBUG") != "" {
		debugFlag = true
	}
	if envFlag = os.Getenv("ENV"); envFlag == "" {
		envFlag = "test"
	}

	wd := os.Getenv("ATHENA_WORK_DIR")
	if wd != "" {
		os.Chdir(wd)
	} else {
		wd, _ = os.Getwd()
	}

	confPath = os.Getenv("ATHENA_CONFIG_PATH")
	if confPath == "" {
		confPath = filepath.Join(wd, "conf")
	}

	logPath = os.Getenv("ATHENA_LOG_PATH")
	if logPath == "" {
		logPath = filepath.Join(wd, "log")
	}
}

func IsDebug() bool {
	return debugFlag
}

func GetEnvFlag() string {
	return envFlag
}

func IsTest() bool {
	return envFlag == "test"
}

func IsOnline() bool {
	return envFlag == "online"
}

func GetConfPath() string {
	return confPath
}

func GetLogPath() string {
	return logPath
}
