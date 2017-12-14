package shared

import (
	"log"
	"os"
	"runtime/debug"
)

func CheckError(err error) {
	if err != nil {
		log.Printf("[ERROR] %s. StackTrace %s", err.Error(), debug.Stack())
		panic(err)
	}
}

func ErrorLog(err error) {
	log.Println("[ERROR] " + err.Error())
}

func DebugLog(text string) {
	log.Println("[DEBUG] " + text)
}

func InfoLog(text string) {
	log.Println("[INFO] " + text)
}

func InitLog(fileName string) {
	f, _ := os.OpenFile(fileName+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	log.SetOutput(f)
}