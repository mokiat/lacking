package main

import (
	"os"

	"github.com/mokiat/lacking/debug/log"
)

func main() {
	log.Info("Started")
	if err := runApplication(); err != nil {
		log.Error("Crashed: %v", err)
		os.Exit(1)
	}
	log.Info("Stopped")
}
