package util

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitSignal(sigs ...os.Signal) {
	c := make(chan os.Signal)
	signal.Notify(c, sigs...)
	<-c
}

func WaitQuitSignal() {
	WaitSignal(syscall.SIGTERM, syscall.SIGINT, syscall.SIGTSTP, syscall.SIGQUIT)
}
