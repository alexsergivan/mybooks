package session

import (
	"sync"
	"time"

	"github.com/alexsergivan/mybooks/repository"

	"github.com/alexsergivan/mybooks/config"
	"github.com/wader/gormstore/v2"
)

var onceStore sync.Once
var store *gormstore.Store

func GetStore() *gormstore.Store {
	onceStore.Do(func() {
		store = gormstore.New(repository.GetDB(), []byte(config.Config("SESSION_SECRET")))
		// db cleanup every hour
		// close quit channel to stop cleanup
		quit := make(chan struct{})
		go store.PeriodicCleanup(3*time.Hour, quit)
	})

	return store
}
