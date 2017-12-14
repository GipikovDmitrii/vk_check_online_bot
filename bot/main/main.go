package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	_ "github.com/mattn/go-sqlite3"
	"../../shared"
)

var (
	//Commands
	start      = "/start"
	addUser    = "➕ Add user"
	removeUser = "➖ Delete user"
	getUsers   = "📋 Get a list of users"
	add        = "➕ Add"
	cancel     = "✖ Cancel"

	selectUserForDelete   = "Select user to delete"
	userSuccessfullyAdded = "User successfully added"
	userNotFound          = "User is not found 😔"

	//Messages
	messageAfterStart   = "You can add a VK user to be notified when the user is online or not online."
	messageAfterAddUser = "Please send a VK ID"

	//Buttons
	addUserButton    = tgbotapi.NewKeyboardButton(addUser)
	removeUserButton = tgbotapi.NewKeyboardButton(removeUser)
	getUsersButton   = tgbotapi.NewKeyboardButton(getUsers)
	addButton        = tgbotapi.NewKeyboardButton(add)
	cancelButton     = tgbotapi.NewKeyboardButton(cancel)

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
			addButton,
		),
		tgbotapi.NewKeyboardButtonRow(
			cancelButton,
		),
	)

	keyboardCancel = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			cancelButton,
		),
	)
)

func main() {
	previewCommands := make(map[int]string)
	previewEnterUser := make(map[int]string)
	shared.InitLog("bot_log")
	bot, err := tgbotapi.NewBotAPI(shared.Token)
	shared.CheckError(err)

	database := shared.InitDB()

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	shared.CheckError(err)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		message := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		userID := update.Message.From.ID
		textMessage := update.Message.Text

		switch textMessage {
		case start:
			shared.AddTelegramUser(database, userID)
			previewCommands[userID] = start
			message.Text = messageAfterStart
			message.ReplyMarkup = keyboardAfterStart
			break
		case addUser:
			previewCommands[userID] = addUser
			message.Text = messageAfterAddUser
			message.ReplyMarkup = keyboardCancel
			break
		case removeUser:
			previewCommands[userID] = removeUser
			users := shared.GetAllVKUserByTelegramUser(database, userID)
			message.Text = selectUserForDelete
			message.ReplyMarkup = getKeyboadrWithAllUsers(users)
			break
		case getUsers:
			previewCommands[userID] = getUsers
			users := shared.GetAllVKUserByTelegramUser(database, userID)
			text := shared.GetFriendlyTextAboutUsers(users)
			if text == "" {
				message.Text = "List is empty."
			} else {
				message.Text = text
			}
			message.ReplyMarkup = keyboardAfterStart
			break
		case add:
			switch previewCommands[userID] {
			case addUser:
				shared.AddVKUser(database, userID, previewEnterUser[userID])
				message.Text = userSuccessfullyAdded
				message.ReplyMarkup = keyboardAfterStart
				previewCommands[userID] = start
			}
			break
		case cancel:
			previewCommands[userID] = start
			message.Text = "Canceled"
			message.ReplyMarkup = keyboardAfterStart
			break
		default:
			switch previewCommands[userID] {
			case addUser:
				user, err := shared.GetUserFromVK(textMessage)
				shared.CheckError(err)
				if user.ID == 0 {
					message.Text = userNotFound
					message.ReplyMarkup = keyboardAfterStart
				} else {
					message.Text = "*" + shared.GetName(user.FirstName, user.LastName) + "*"
					previewEnterUser[userID] = textMessage
					message.ReplyMarkup = keyboardAfterEnterUser
					message.ParseMode = "Markdown"
				}
				break
			case removeUser:
				shared.RemoveVKUser(database, textMessage, userID)
				message.Text = "User *" + textMessage + "* successfully deleted"
				message.ReplyMarkup = keyboardAfterStart
				message.ParseMode = "Markdown"
				break
			}
		}

		log.Printf("[%s] %s | preview message: %s", update.Message.From.UserName, update.Message.Text, previewCommands[userID])

		bot.Send(message)
	}
}

func getKeyboadrWithAllUsers(users map[string]string) tgbotapi.ReplyKeyboardMarkup {
	buttons := make([][]tgbotapi.KeyboardButton, len(users)+1)
	i := 0
	for k, v := range users {
		buttons[i] = tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(v + " [" + k + "]"))
		i++
	}
	buttons[i] = tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(cancel))
	keyboard := tgbotapi.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard:       buttons,
	}
	return keyboard
}