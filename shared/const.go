package shared

var (
	Token            = ""
	TelegramURL      = "https://api.telegram.org/bot"
	DataBaseFilePath = "../../database/vknotification.db"
	AccessTokenVk    = ""
	vkURL            = "https://api.vk.com/method/users.get"
	APIVersion       = "5.69"
	Fields           = "online,last_seen"

	//SQL
	GetAllUsersQuery               = "SELECT vk.id, tel.telegram_id, vk.last_online, vk.last_platform, vk.vk_id FROM vk_user vk LEFT JOIN telegram_user tel ON vk.telegram_user_id = tel.id"
	UpdateUserQuery                = "UPDATE vk_user SET last_online = ?, last_platform = ? WHERE id = ?"
	AddTelegramUserQuery           = "INSERT INTO telegram_user (telegram_id) VALUES (?)"
	AddVKUserQuery                 = "INSERT INTO vk_user (telegram_user_id, vk_id, last_online, last_platform) VALUES ((SELECT telegram_user.id FROM telegram_user WHERE telegram_id = ?), ?, 0, 0)"
	RemoveVKUserQuery              = "DELETE FROM vk_user WHERE vk_id = ? AND telegram_user_id = (SELECT id FROM telegram_user WHERE telegram_id = ?)"
	GetAllUsersByTelegramUserQuery = "SELECT vk_id FROM vk_user WHERE telegram_user_id = (SELECT id FROM telegram_user WHERE telegram_id = ?)"

	PlatformMap = map[int]string{
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
