package resolvers

import (
	"net/http"
	"time"

	"github.com/alexsergivan/mybooks/cache"

	"github.com/alexsergivan/mybooks/userbook"
	"github.com/labstack/echo/v4"
	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"
)

func GetLibrary(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		ristrettoCache := cache.NewRistrettoCache()
		cacheKey := "library"
		templateData := map[string]interface{}{}
		library, found := ristrettoCache.Get(cacheKey)
		if !found {
			books := userbook.GetBooksListGroupedByLetter(db)
			ristrettoCache.Set(cacheKey, books, time.Hour*24)
			time.Sleep(10 * time.Millisecond)
			templateData = map[string]interface{}{
				"books": books,
			}
		} else {
			templateData = map[string]interface{}{
				"books": library.(map[string][]userbook.Book),
			}
		}

		return c.Render(http.StatusOK, "books--library", templateData)
	}
}
