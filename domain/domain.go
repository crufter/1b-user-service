package domain

import (
	"time"

	shortid "github.com/ventu-io/go-shortid"
)

var Sid *shortid.Shortid

func init() {
	var err error
	Sid, err = shortid.New(1, shortid.DefaultABC, 2719)
	if err != nil {
		panic(err)
	}
}

type User struct {
	Id         string
	Password   string `json:"-"`
	Nick       string
	Name       string
	Email      string
	AvatarLink string
	Premium    bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// For logins
type AccessToken struct {
	Id        string
	Token     string
	UserId    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Reset struct {
	Id        string
	Secret    string
	UserId    string
	CreatedAt time.Time
	Used      bool
}
