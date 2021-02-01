package pack

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type Context struct {
	logMutex  *sync.Mutex
	logDepth  int
	logLines  []string
	logBuffer bytes.Buffer

	storageMutex *sync.Mutex
	storage      Storage
}

func (c *Context) LogAction(description string) func() {
	startTime := time.Now()

	padding := strings.Repeat("    ", c.logDepth)
	c.log("%s--> %s", padding, description)
	c.logDepth++

	return func() {
		c.log("%s<-- %s (%s)", padding, description, time.Since(startTime))
		if c.logDepth--; c.logDepth == 0 {
			c.flushLog()
		}
	}
}

func (c *Context) IO(handler func(Storage) error) error {
	c.storageMutex.Lock()
	defer c.storageMutex.Unlock()
	return handler(c.storage)
}

func (c *Context) log(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	c.logLines = append(c.logLines, line)
}

func (c *Context) flushLog() {
	c.logMutex.Lock()
	defer c.logMutex.Unlock()

	for _, line := range c.logLines {
		log.Println(line)
	}
	log.Println()
}
