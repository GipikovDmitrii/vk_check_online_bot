package shared

import (
	"bytes"
	"strconv"
)

func GetName(firstName string, lastName string) string {
	return firstName + " " + lastName
}

func GetWebPlatform(platform int) string {
	return PlatformMap[platform]
}

func GetDiff(currentUser UserDataBase, actualUser User) (string, bool) {
	result := ""
	diff := false
	if currentUser.LastPlatform != actualUser.LastSeen.Platform {
		result = " changed platform, current: " + GetWebPlatform(actualUser.LastSeen.Platform)
		diff = true
	}
	if currentUser.LastOnline == 0 && actualUser.Online == 1 {
		result = " is online. Current platform: " + GetWebPlatform(actualUser.LastSeen.Platform)
		diff = true
	}
	if currentUser.LastOnline == 1 && actualUser.Online == 0 {
		result = " is offline. Last platform: " + GetWebPlatform(actualUser.LastSeen.Platform)
		diff = true
	}
	result = GetName(actualUser.FirstName, actualUser.LastName) + result
	return result, diff
}

func IsOnline(online int) string {
	if online == 1 {
		return "YES"
	} else {
		return "NO"
	}
}

func GetFriendlyTextAboutUsers(users map[string]string) string {
	var buffer bytes.Buffer
	for k := range users {
		buffer.WriteString(GetFriendlyTextAboutUser(k))
		buffer.WriteString("\n\n")
	}
	return buffer.String()
}

func GetFriendlyTextAboutUser(userId string) string {
	user, e := GetUserFromVK(userId)
	CheckError(e)
	var buffer bytes.Buffer
	buffer.WriteString("ID: " + strconv.Itoa(user.ID) + "\n")
	buffer.WriteString("Name: " + GetName(user.FirstName, user.LastName) + "\n")
	buffer.WriteString("Online: " + IsOnline(user.Online) + "\n")
	buffer.WriteString("Platform: " + GetWebPlatform(user.LastSeen.Platform))
	return buffer.String()
}
