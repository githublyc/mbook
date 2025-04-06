package domain

import "time"

type User struct {
	Id         int64
	Email      string
	Password   string
	Ctime      time.Time
	Nickname   string
	Birthday   time.Time
	AboutMe    string
	Phone      string
	WeChatInfo WeChatInfo
}
