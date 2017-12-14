package shared

import (
	"database/sql"
	"fmt"
	"strings"
	"log"
)

func GetAllVKUsers(database *sql.DB) []UserDataBase {
	var users []UserDataBase
	rows, e := database.Query(GetAllUsersQuery)
	CheckError(e)
	defer rows.Close()
	for rows.Next() {
		var id string
		var telegramUserId string
		var lastOnline int
		var lastPlatform int
		var vkId string
		err := rows.Scan(&id, &telegramUserId, &lastOnline, &lastPlatform, &vkId)
		CheckError(err)
		users = append(users, UserDataBase{id, telegramUserId, lastOnline, lastPlatform, vkId})
	}
	return users
}

func AddVKUser(database *sql.DB, telegramID int, vkID string) {
	statement, e := database.Prepare(AddVKUserQuery)
	CheckError(e)
	_, e = statement.Exec(telegramID, vkID)
	if e != nil && strings.Contains(e.Error(), "UNIQUE") {
		log.Printf("VK user already exists")
	} else if e != nil && !strings.Contains(e.Error(), "UNIQUE") {
		ErrorLog(e)
	} else {
		InfoLog(fmt.Sprintf("VK user [%d] added for user: %d.", vkID, telegramID))
	}
}

func UpdateVKUsers(database *sql.DB, id string, lastOnline int, lastPlatform int) {
	statement, e := database.Prepare(UpdateUserQuery)
	CheckError(e)
	_, e = statement.Exec(lastOnline, lastPlatform, id)
	if e != nil {
		ErrorLog(e)
	} else {
		InfoLog(fmt.Sprintf("VK user [%s] updated.", id))
	}
}

func GetAllVKUserByTelegramUser(database *sql.DB, telegramID int) map[string]string {
	rows, e := database.Query(GetAllUsersByTelegramUserQuery, telegramID)
	CheckError(e)
	users := make(map[string]string)
	defer rows.Close()
	for rows.Next() {
		var vkId string
		err := rows.Scan(&vkId)
		CheckError(err)
		user, err := GetUserFromVK(vkId)
		CheckError(err)
		users[vkId] = GetName(user.FirstName, user.LastName)
	}
	return users
}

func RemoveVKUser(database *sql.DB, vkID string, telegramID int) {
	statement, e := database.Prepare(RemoveVKUserQuery)
	CheckError(e)
	startIndex := strings.Index(vkID, "[")
	finishIndex := strings.Index(vkID, "]")
	vkID = vkID[startIndex+1:finishIndex]
	_, e = statement.Exec(vkID, telegramID)
	if e != nil {
		ErrorLog(e)
	} else {
		InfoLog(fmt.Sprintf("VK user [%d] remove for user: %d.", vkID, telegramID))
	}
}

func AddTelegramUser(database *sql.DB, telegramID int) {
	statement, e := database.Prepare(AddTelegramUserQuery)
	CheckError(e)
	_, e = statement.Exec(telegramID)
	if e != nil && strings.Contains(e.Error(), "UNIQUE") {
		log.Printf("Telegram user already exists")
	} else if e != nil && !strings.Contains(e.Error(), "UNIQUE") {
		ErrorLog(e)
	} else {
		InfoLog(fmt.Sprintf("Telegram user [%d] added.", telegramID))
	}
}

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", DataBaseFilePath)
	CheckError(err)
	return db
}
