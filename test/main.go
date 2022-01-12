package main

import (
	"context"

	"github.com/rickslab/ares/logger"
)

func main() {
	logger.Init()

	entry := logger.NewEntry(context.Background(), nil)
	entry.Info("Hello")
}
