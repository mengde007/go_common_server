package main

import (
	"logger"
	mj "majiangserver"
)

func main() {
	logger.Info("Test Start")
	mj.TestBao()
	logger.Info("Test End")
}
