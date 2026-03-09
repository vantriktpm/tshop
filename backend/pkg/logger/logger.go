package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// entry is the JSON structure written for each log line.
type entry struct {
	Level   string                 `json:"level"`
	Time    string                 `json:"time"`
	Message string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

var stdLogger = log.New(os.Stdout, "", 0)

// logWithLevel writes a single JSON log line to stdout.
func logWithLevel(level, msg string, fields map[string]interface{}) {
	e := entry{
		Level:   level,
		Time:    time.Now().UTC().Format(time.RFC3339Nano),
		Message: msg,
		Fields:  fields,
	}
	b, err := json.Marshal(e)
	if err != nil {
		stdLogger.Printf(`{"level":"ERROR","time":"%s","message":"logger_marshal_failed","fields":{"err":"%v"}}`,
			time.Now().UTC().Format(time.RFC3339Nano), err)
		return
	}
	stdLogger.Println(string(b))
}

// Info logs an informational message.
func Info(msg string, fields map[string]interface{}) {
	logWithLevel("INFO", msg, fields)
}

// Error logs an error message.
func Error(msg string, fields map[string]interface{}) {
	logWithLevel("ERROR", msg, fields)
}

// Debug logs a debug message (can be filtered at ELK level).
func Debug(msg string, fields map[string]interface{}) {
	logWithLevel("DEBUG", msg, fields)
}

