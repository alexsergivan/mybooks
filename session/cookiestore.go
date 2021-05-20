package session

import (
	"sync"

	"github.com/alexsergivan/mybooks/config"

	"github.com/gorilla/sessions"
)

var cookieStore *sessions.CookieStore

var onceCokieStore sync.Once

func GetCookieStore() *sessions.CookieStore {
	onceCokieStore.Do(func() {
		cookieStore = sessions.NewCookieStore([]byte(config.Config("SESSION_SECRET")))
	})

	return cookieStore
}
