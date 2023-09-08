package util

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitSignal(sigs ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sigs...)
	<-c
}

func WaitQuitSignal() {
	WaitSignal(syscall.SIGTERM, syscall.SIGINT, syscall.SIGTSTP, syscall.SIGQUIT)
}
