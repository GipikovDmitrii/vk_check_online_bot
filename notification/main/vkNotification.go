package main

import (
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
	"log"
	_ "github.com/mattn/go-sqlite3"
	"../../shared"
)

func main() {
	database := shared.InitDB()
	for {
		usersList := shared.GetAllVKUsers(database)
		for _, currentUser := range usersList {
			actualUser, err := shared.GetUserFromVK(currentUser.VkID)
			shared.CheckError(err)
			text, diff := shared.GetDiff(currentUser, actualUser)
			if diff {
				shared.UpdateVKUsers(database, currentUser.ID, actualUser.Online, actualUser.LastSeen.Platform)
				sendNotification(currentUser.TelegramUserId, text)
			}
		}
		time.Sleep(time.Second * 30)
	}
}

func sendNotification(telegramId string, text string) {
	values := url.Values{}
	values.Set("chat_id", telegramId)
	values.Set("text", text)
	uri, _ := url.Parse(shared.TelegramURL + shared.Token + "/sendMessage")
	uri.RawQuery = values.Encode()
	shared.DebugLog("URL for notification: " + uri.String())
	response, e := http.Get(uri.String())
	if e != nil {
		shared.CheckError(e)
	}
	message := &shared.ResponseTelegram{
		Status:      false,
		Description: ""}
	body, e := ioutil.ReadAll(response.Body)
	if e != nil {
		shared.CheckError(e)
	}
	e = json.Unmarshal([]byte(body), message)
	if e != nil {
		shared.CheckError(e)
	}
	if message.Status {
		shared.InfoLog("Message sended. Text: " + text)
	} else {
		log.Println("[ERROR] Message not send. Description: " + message.Description)
	}
}
