package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"encoding/json"
	"net/http"
	"net/url"
	"io/ioutil"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"bytes"
	"strconv"
)

var (
	token         = ""
	accessTokenVk = ""
	urlVk         = "https://api.vk.com/method/users.get"
	versionApi    = "5.69"
	fields        = "online,last_seen"

	//SQL
	addTelegramUserQuery           = "INSERT INTO telegram_user (telegram_id) VALUES (?)"
	addVKUserQuery                 = "INSERT INTO vk_user (telegram_user_id, vk_id, last_online, last_platform) VALUES ((SELECT telegram_user.id FROM telegram_user WHERE telegram_id = ?), ?, 0, 0)"
	removeVKUserQuery              = "DELETE FROM vk_user WHERE vk_id = ? AND telegram_user_id = (SELECT id FROM telegram_user WHERE telegram_id = ?)"
	getAllUsersByTelegramUserQuery = "SELECT vk_id FROM vk_user WHERE telegram_user_id = (SELECT id FROM telegram_user WHERE telegram_id = ?)"

	//Commands
	start      = "/start"
	addUser    = "Add user"
	removeUser = "Remove user"
	getUsers   = "Get users"
	OK         = "Имя правильное. Добавить"
	NO         = "Имя не правильное"

	//Messages
	messageAfterStart   = "Привет. Ты можешь добавить пользователя ВК, чтобы получать уведомления о том, когда пользователь появился или исчез из онлайна."
	messageAfterAddUser = "Пожалуйста, пришлите ID пользователя."

	//Buttons
	addUserButton    = tgbotapi.NewKeyboardButton(addUser)
	removeUserButton = tgbotapi.NewKeyboardButton(removeUser)
	getUsersButton   = tgbotapi.NewKeyboardButton(getUsers)
	OKButton         = tgbotapi.NewKeyboardButton(OK)
	NOButton         = tgbotapi.NewKeyboardButton(NO)

	//Keyboards
	keyboardAfterStart = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			addUserButton,
		),
		tgbotapi.NewKeyboardButtonRow(
			removeUserButton,
		),
		tgbotapi.NewKeyboardButtonRow(
			getUsersButton,
		),
	)

	keyboardAfterEnterUser = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			OKButton,
		),
		tgbotapi.NewKeyboardButtonRow(
			NOButton,
		),
	)
)

func main() {
	previewCommands := make(map[int]string)
	previewEnterUser := make(map[int]string)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	database := InitDB("vknotification.db")

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		message := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		userID := update.Message.From.ID
		textMessage := update.Message.Text

		switch textMessage {
		case start:
			addTelegramUser(database, userID)
			previewCommands[userID] = start
			message.Text = messageAfterStart
			message.ReplyMarkup = keyboardAfterStart
			log.Println(keyboardAfterStart)
			break
		case addUser:
			previewCommands[userID] = addUser
			message.Text = messageAfterAddUser
			message.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			break
		case removeUser:
			previewCommands[userID] = removeUser
			users := getAllVKUserByTelegramUser(database, userID)
			message.Text = "Выберите пользователя для удаления"
			message.ReplyMarkup = getKeyboadrWithAllUsers(users)
			break
		case getUsers:
			previewCommands[userID] = getUsers
			users := getAllVKUserByTelegramUser(database, userID)
			message.Text = getFriendlyTextAboutUsers(users)
			message.ReplyMarkup = keyboardAfterStart
			break
		case OK:
			switch previewCommands[userID] {
			case addUser:
				addVKUser(database, userID, previewEnterUser[userID])
				message.Text = "Пользователь добавлен"
				message.ReplyMarkup = keyboardAfterStart
				previewCommands[userID] = start
			}
			break
		case NO:
			switch previewCommands[userID] {
			case addUser:
				message.Text = "К сожалению, мы не нашли пользователя"
				message.ReplyMarkup = keyboardAfterStart
				previewCommands[userID] = start
				break
			}
			break
		default:
			switch previewCommands[userID] {
			case addUser:
				user, err := getUser(textMessage)
				errorLog(err)
				previewEnterUser[userID] = textMessage
				message.Text = "*" + getName(user.FirstName, user.LastName) + "*"
				message.ReplyMarkup = keyboardAfterEnterUser
				message.ParseMode = "Markdown"
				break
			case removeUser:
				removeVKUser(database, textMessage, userID)
				message.Text = "Пользователь *" + textMessage + "* удален из списка"
				message.ReplyMarkup = keyboardAfterStart
				message.ParseMode = "Markdown"
				break
			}
		}

		log.Printf("[%s] %s | preview message: %s", update.Message.From.UserName, update.Message.Text, previewCommands[userID])

		bot.Send(message)
	}
}

func getName(firstName string, lastName string) string {
	return firstName + " " + lastName
}

func isOnline(online int) string {
	if online == 1 {
		return "yes"
	} else {
		return "no"
	}
}

func isWebPlatform(platform int) string {
	if platform == 7 {
		return "yes"
	} else {
		return "no"
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
	return message.User[0], nil
}

func getFriendlyTextAboutUsers(users map[string]string) string {
	var buffer bytes.Buffer
	for k := range users {
		buffer.WriteString(getFriendlyTextAboutUser(k))
		buffer.WriteString("\n\n")
	}
	return buffer.String()
}

func getFriendlyTextAboutUser(userId string) string {
	user, error := getUser(userId)
	errorLog(error)
	var buffer bytes.Buffer
	buffer.WriteString("ID: " + strconv.Itoa(user.ID) + "\n")
	buffer.WriteString("Name: " + getName(user.FirstName, user.LastName) + "\n")
	buffer.WriteString("Online: " + isOnline(user.Online) + "\n")
	buffer.WriteString("Web: " + isWebPlatform(user.LastSeen.Platform))
	return buffer.String()
}

func addTelegramUser(database *sql.DB, telegramID int) {
	statement, e := database.Prepare(addTelegramUserQuery)
	errorLog(e)
	result, e := statement.Exec(telegramID)
	if e != nil && strings.Contains(e.Error(), "UNIQUE") {
		log.Printf("Telegram user already exists")
	} else if e != nil && !strings.Contains(e.Error(), "UNIQUE") {
		log.Panic(e)
	} else {
		log.Printf("Telegram user [%d] added. Result: %s", telegramID, result)
	}
}

func addVKUser(database *sql.DB, telegramID int, vkID string) {
	statement, e := database.Prepare(addVKUserQuery)
	errorLog(e)
	result, e := statement.Exec(telegramID, vkID)
	if e != nil && strings.Contains(e.Error(), "UNIQUE") {
		log.Printf("VK user already exists")
	} else if e != nil && !strings.Contains(e.Error(), "UNIQUE") {
		log.Panic(e)
	} else {
		log.Printf("VK user [%d] added for user: %d. Result: %s", vkID, telegramID, result)
	}
}

func removeVKUser(database *sql.DB, vkID string, telegramID int) {
	statement, e := database.Prepare(removeVKUserQuery)
	errorLog(e)
	startIndex := strings.Index(vkID, "[")
	finishIndex := strings.Index(vkID, "]")
	vkID = vkID[startIndex+1:finishIndex]
	log.Println(vkID)
	result, e := statement.Exec(vkID, telegramID)
	if e != nil {
		log.Panic(e)
	} else {
		log.Printf("VK user [%d] remove for user: %d. Result: %s", vkID, telegramID, result)
	}
}

func getAllVKUserByTelegramUser(database *sql.DB, telegramID int) map[string]string {
	log.Println(telegramID)
	rows, e := database.Query(getAllUsersByTelegramUserQuery, telegramID)
	errorLog(e)
	users := make(map[string]string)
	defer rows.Close()
	for rows.Next() {
		var vkId string
		err := rows.Scan(&vkId)
		errorLog(err)
		user, err := getUser(vkId)
		errorLog(err)
		users[vkId] = getName(user.FirstName, user.LastName)
	}
	return users
}

func getKeyboadrWithAllUsers(users map[string]string) tgbotapi.ReplyKeyboardMarkup {
	buttons := make([][]tgbotapi.KeyboardButton, len(users))
	i := 0
	for k, v := range users {
		buttons[i] = tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(v + " [" + k + "]"))
		i++
	}
	keyboard := tgbotapi.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard:       buttons,
	}
	return keyboard
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

func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	errorLog(err)
	return db
}

func errorLog(err error) {
	if err != nil {
		panic(err)
	}
}
