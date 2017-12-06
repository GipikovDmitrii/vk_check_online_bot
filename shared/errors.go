package shared

import "log"

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func ErrorLog(err error)  {
	log.Println("[ERROR] " + err.Error())
}

func DebugLog(text string)  {
	log.Println("[DEBUG] " + text)
}

func InfoLog(text string)  {
	log.Println("[INFO] " + text)
}
