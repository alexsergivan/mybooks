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
		letter := c.QueryParam("letter")
		if letter == "" {
			templateData := map[string]interface{}{}
			cacheKey := "alphabet"
			alphabet, found := ristrettoCache.Get(cacheKey)
			if !found {
				alphabet := userbook.GetAlphabet(db)
				ristrettoCache.Set(cacheKey, alphabet, time.Hour*24)
				time.Sleep(10 * time.Millisecond)
				templateData = map[string]interface{}{
					"alphabet":   alphabet,
					"categories": userbook.GetBookCategories(db),
				}
			} else {
				templateData = map[string]interface{}{
					"alphabet":   alphabet.([]string),
					"categories": userbook.GetBookCategories(db),
				}
			}
			return c.Render(http.StatusOK, "books--alphabet", templateData)

		} else {

			cacheKey := "library_" + letter
			templateData := map[string]interface{}{}
			library, found := ristrettoCache.Get(cacheKey)
			if !found {
				books := userbook.GetBooksListGroupedByLetter(db, letter)
				ristrettoCache.Set(cacheKey, books, time.Hour*24)
				time.Sleep(10 * time.Millisecond)
				templateData = map[string]interface{}{
					"books":  books,
					"letter": letter,
				}
			} else {
				templateData = map[string]interface{}{
					"books":  library.(map[string][]userbook.Book),
					"letter": letter,
				}
			}

			return c.Render(http.StatusOK, "books--library", templateData)
		}
	}
}
