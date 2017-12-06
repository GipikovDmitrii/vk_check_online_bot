package shared

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