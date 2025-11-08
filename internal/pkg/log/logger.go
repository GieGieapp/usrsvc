package log

import (
	"log"
	"os"
)

var (
	Info  = log.New(os.Stdout, "INFO ", log.LstdFlags|log.Lmicroseconds)
	Error = log.New(os.Stderr, "ERROR ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
)
