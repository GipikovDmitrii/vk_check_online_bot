package main

import (
	"database/sql"
	"net/url"
	"bytes"
	"strconv"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
	"log"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
)

var (
	token         = ""
	accessTokenVk = ""
	urlVk         = "https://api.vk.com/method/users.get"
	urlTelegram   = "https://api.telegram.org/bot"
	versionApi    = "5.69"
	fields        = "online,last_seen"

	//SQL
	getAllUsersQuery = "SELECT vk.id, tel.telegram_id, vk.last_online, vk.last_platform, vk.vk_id FROM vk_user AS vk, telegram_user AS tel"
	updateUserQuery  = "UPDATE vk_user SET last_online = ?, last_platform = ? WHERE id = ?"

	platfotmMap = map[int]string{
		1: "Mobile version",
		2: "iPhone",
		3: "iPad",
		4: "Android",
		5: "Windows Phone",
		6: "Windows 10",
		7: "PC browser",
		8: "VK Mobile",
	}
)

func main() {
	database := InitDB("./database/vknotification.db")
	for {
		usersList := getAllVKUsers(database)
		for _, currentUser := range usersList {
			actualUser, err := getUser(currentUser.VkID)
			checkError(err)
			text, diff := getDiff(currentUser, actualUser)
			if diff {
				updateVKUsers(database, currentUser.ID, actualUser.Online, actualUser.LastSeen.Platform)
				sendNotification(currentUser.TelegramUserId, text)
			}
		}
		time.Sleep(time.Second * 10)
	}
}

func getDiff(currentUser UserDataBase, actualUser User) (string, bool) {
	result := ""
	diff := false
	if currentUser.LastPlatform != actualUser.LastSeen.Platform {
		result = " changed platform, current: " + getWebPlatform(actualUser.LastSeen.Platform)
		diff = true
	}
	if currentUser.LastOnline == 0 && actualUser.Online == 1 {
		result = " is online. Current platform: " + getWebPlatform(actualUser.LastSeen.Platform)
		diff = true
	}
	if currentUser.LastOnline == 1 && actualUser.Online == 0 {
		result = " is offline. Last platform: " + getWebPlatform(actualUser.LastSeen.Platform)
		diff = true
	}
	result = getName(actualUser.FirstName, actualUser.LastName) + result
	return result, diff
}

func sendNotification(telegramId string, text string) {
	values := url.Values{}
	values.Set("chat_id", telegramId)
	values.Set("text", text)
	uri, _ := url.Parse(urlTelegram + token + "/sendMessage")
	uri.RawQuery = values.Encode()
	debugLog("URL for notification: " + uri.String())
	response, e := http.Get(uri.String())
	if e != nil {
		checkError(e)
	}
	message := &ResponseTelegram{
		Status:      false,
		Description: ""}
	body, e := ioutil.ReadAll(response.Body)
	if e != nil {
		checkError(e)
	}
	e = json.Unmarshal([]byte(body), message)
	if e != nil {
		checkError(e)
	}
	if message.Status {
		infoLog("Message sended. Text: " + text)
	} else {
		log.Println("[ERROR] Message not send. Description: " + message.Description)
	}
}

func getAllVKUsers(database *sql.DB) []UserDataBase {
	var users []UserDataBase
	rows, e := database.Query(getAllUsersQuery)
	checkError(e)
	defer rows.Close()
	for rows.Next() {
		var id string
		var telegramUserId string
		var lastOnline int
		var lastPlatform int
		var vkId string
		err := rows.Scan(&id, &telegramUserId, &lastOnline, &lastPlatform, &vkId)
		checkError(err)
		users = append(users, UserDataBase{id, telegramUserId, lastOnline, lastPlatform, vkId})
	}
	return users
}

func updateVKUsers(database *sql.DB, id string, lastOnline int, lastPlatform int) {
	statement, e := database.Prepare(updateUserQuery)
	checkError(e)
	result, e := statement.Exec(lastOnline, lastPlatform, id)
	if e != nil {
		errorLog(e)
	} else {
		infoLog(fmt.Sprintf("VK user [%s] updated. Result: %s", id, result))
	}
}

func getUser(userIds string) (User, error) {
	response, e := http.Get(getURL(userIds))
	if e != nil {
		return User{}, e
	}
	message := &Response{
		User: []User{}}
	body, e := ioutil.ReadAll(response.Body)
	if e != nil {
		return User{}, e
	}
	e = json.Unmarshal([]byte(body), message)
	if e != nil {
		return User{}, e
	}
	if len(message.User) == 0 {
		return User{}, e
	} else {
		return message.User[0], nil
	}
}

func getFriendlyTextAboutUser(userId string) string {
	user, e := getUser(userId)
	checkError(e)
	var buffer bytes.Buffer
	buffer.WriteString("ID: " + strconv.Itoa(user.ID) + "\n")
	buffer.WriteString("Name: " + getName(user.FirstName, user.LastName) + "\n")
	buffer.WriteString("Online: " + isOnline(user.Online) + "\n")
	buffer.WriteString("Platform: " + getWebPlatform(user.LastSeen.Platform))
	return buffer.String()
}

func getName(firstName string, lastName string) string {
	return firstName + " " + lastName
}

func isOnline(online int) string {
	if online == 1 {
		return "YES"
	} else {
		return "NO"
	}
}

func getWebPlatform(platform int) string {
	return platfotmMap[platform]
}

func getURL(userIds string) string {
	values := url.Values{}
	values.Set("access_token", accessTokenVk)
	values.Set("user_ids", userIds)
	values.Set("fields", fields)
	values.Set("v", versionApi)
	uri, _ := url.Parse(urlVk)
	uri.RawQuery = values.Encode()
	return uri.String()
}

type ResponseTelegram struct {
	Status      bool   `json:"ok"`
	Description string `json:"description"`
}

type Response struct {
	User []User `json:"response"`
}

type User struct {
	ID        int      `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Online    int      `json:"online"`
	LastSeen  LastSeen `json:"last_seen"`
}

type LastSeen struct {
	Time     int `json:"time"`
	Platform int `json:"platform"`
}

type UserDataBase struct {
	ID             string `json:"id"`
	TelegramUserId string `json:"telegram_user_id"`
	LastOnline     int    `json:"last_online"`
	LastPlatform   int    `json:"last_platform"`
	VkID           string `json:"vk_id"`
}

func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	checkError(err)
	return db
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func errorLog(err error)  {
	log.Println("[ERROR] " + err.Error())
}

func debugLog(text string)  {
	log.Println("[DEBUG] " + text)
}

func infoLog(text string)  {
	log.Println("[INFO] " + text)
}