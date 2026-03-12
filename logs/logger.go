package logs

import (
	"fmt"
	"log"
	"os"
	"time"
)

var logFile *os.File

func Init() {
	os.MkdirAll("./logs", os.ModePerm)
	var err error
	logFile, err = os.OpenFile("./logs/middleware.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Impossible d'ouvrir le fichier de logs:", err)
	}
}

func write(level, message string) {
	entry := fmt.Sprintf("[%s] [%s] %s",
		time.Now().Format("2006-01-02 15:04:05"), level, message)
	fmt.Println(entry)
	if logFile != nil {
		logFile.WriteString(entry + "\n")
	}
}

func Info(msg string)    { write("INFO", msg) }
func Warning(msg string) { write("WARNING", msg) }
func Error(msg string)   { write("ERROR", msg) }
func Success(msg string) { write("SUCCESS", msg) }
