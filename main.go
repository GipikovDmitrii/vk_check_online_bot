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
)

var (
	token         = ""
	accessTokenVk = ""
	urlVk         = "https://api.vk.com/method/users.get"
	versionApi    = "5.69"
	fields        = "online,last_seen"

	//SQL
	addTelegramUserSQL = "INSERT INTO telegram_user (telegram_id) VALUES (?)"
	addVKUserSQL       = "INSERT INTO vk_user (telegram_user_id, vk_id, last_online, last_platform) VALUES (?, ?, 0, 0)"
	removeVKUserSQL    = "DELETE FROM vk_user WHERE vk_id = ? AND telegram_user_id = (SELECT id FROM telegram_user WHERE telegram_id = ?)"

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
	prevuesCommands := make(map[int]string)
	prevuesEnterUser := make(map[int]string)

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
			prevuesCommands[userID] = start
			message.Text = messageAfterStart
			message.ReplyMarkup = keyboardAfterStart
			break
		case addUser:
			prevuesCommands[userID] = addUser
			message.Text = messageAfterAddUser
			message.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			break
		case removeUser:
			prevuesCommands[userID] = removeUser
			break
		case getUsers:
			prevuesCommands[userID] = getUsers
			break
		case OK:
			switch prevuesCommands[userID] {
			case addUser:
				addVKUser(database, userID, prevuesEnterUser[userID])
				message.Text = "Пользователь добавлен"
				message.ReplyMarkup = keyboardAfterStart
				prevuesCommands[userID] = start
			}
			break
		case NO:
			switch prevuesCommands[userID] {
			case addUser:
				message.Text = "К сожалению, мы не нашли пользователя"
				message.ReplyMarkup = keyboardAfterStart
				prevuesCommands[userID] = start
			}
			break
		default:
			switch prevuesCommands[userID] {
			case addUser:
				user, err := getUser(textMessage)
				if err != nil {
					log.Panic(err)
				}
				prevuesEnterUser[userID] = textMessage
				message.Text = "*" + getName(user.FirstName, user.LastName) + "*"
				message.ReplyMarkup = keyboardAfterEnterUser
				message.ParseMode = "Markdown"
				break
			}
		}

		log.Printf("[%s] %s | preview message: %s", update.Message.From.UserName, update.Message.Text, prevuesCommands[userID])

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

func addTelegramUser(database *sql.DB, telegramID int) {
	statement, e := database.Prepare(addTelegramUserSQL)
	if e != nil {
		log.Panic(e)
	}
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
	statement, e := database.Prepare(addVKUserSQL)
	if e != nil {
		log.Panic(e)
	}
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
	statement, e := database.Prepare(removeVKUserSQL)
	if e != nil {
		log.Panic(e)
	}
	result, e := statement.Exec(vkID, telegramID)
	if e != nil {
		log.Panic(e)
	} else {
		log.Printf("VK user [%d] remove for user: %d. Result: %s", vkID, telegramID, result)
	}
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
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("db nil")
	}
	return db
}
