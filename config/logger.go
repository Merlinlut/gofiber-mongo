package config

import (
	"log"
	"os"
)

func GetLogger() *log.Logger {
	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}
