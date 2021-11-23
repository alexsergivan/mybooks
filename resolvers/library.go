package resolvers

import (
	"net/http"

	"github.com/alexsergivan/mybooks/userbook"
	"github.com/labstack/echo/v4"
	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"
)

func GetLibrary(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		books := userbook.GetBooksListGroupedByLetter(db)
		templateData := map[string]interface{}{
			"books": books,
		}

		return c.Render(http.StatusOK, "books--library", templateData)
	}
}
