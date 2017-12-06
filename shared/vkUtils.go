package shared

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"net/url"
)

func GetUserFromVK(userIds string) (User, error) {
	response, e := http.Get(GetURLForVK(userIds))
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

func GetURLForVK(userIds string) string {
	values := url.Values{}
	values.Set("access_token", AccessTokenVk)
	values.Set("user_ids", userIds)
	values.Set("fields", Fields)
	values.Set("v", APIVersion)
	uri, _ := url.Parse(vkURL)
	uri.RawQuery = values.Encode()
	return uri.String()
}
