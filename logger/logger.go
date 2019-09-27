package logger

import (
	"log"
	"os"
)

var ProgressLogger = log.New(os.Stdout, "goweasyprint.progress ", log.LstdFlags)
