package log

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func formatLog(level string, format string, args ...interface{}) string {
	return fmt.Sprintf("[%s] [%s] %s", time.Now().UTC().Format(time.RFC3339), level, fmt.Sprintf(format, args...))
}

func Info(format string, args ...interface{}) {
	log.Println(formatLog("INFO", format, args...))
}

func Warn(format string, args ...interface{}) {
	log.Println(formatLog("WARN", format, args...))
}

func Error(format string, args ...interface{}) {
	log.Println(formatLog("ERROR", format, args...))
}

func Debug(format string, args ...interface{}) {
	log.Println(formatLog("DEBUG", format, args...))
}

func Panic(format string, args ...interface{}) {
	log.Panicln(formatLog("PANIC", format, args...))
}

func Panicf(format string, args ...interface{}) {
	log.Panicf("%s", formatLog("PANIC", format, args...))
}

func Fatal(format string, args ...interface{}) {
	log.Fatal(formatLog("FATAL", format, args...))
}

func LogHTTPError(c *gin.Context, err error, status int) {
	log.Printf("[%s] [HTTP_ERROR] %s %s %d %s - remote=%s", time.Now().UTC().Format(time.RFC3339), c.Request.Method, c.Request.URL.String(), status, err.Error(), c.ClientIP())
}
