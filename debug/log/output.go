package log

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	bufferMU sync.Mutex
	buffer   bytes.Buffer

	goLogger = log.New(os.Stderr, "", log.LstdFlags|log.Lmsgprefix)
)

func output(level, padding, namespace, format string, args ...any) {
	bufferMU.Lock()
	defer bufferMU.Unlock()

	defer buffer.Reset()
	fmt.Fprintf(&buffer, "[ %s ]%s [ %s ] ", level, padding, namespace)
	fmt.Fprintf(&buffer, format, args...)

	goLogger.Output(4, buffer.String())
}
