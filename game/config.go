package game

import "time"

type AppConfig struct {
	WindowTitle        string
	WindowWidth        int
	WindowHeight       int
	WindowHideCursor   bool
	WindowVSync        bool
	UpdateLoopInterval time.Duration
}
