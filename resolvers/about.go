package resolvers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"
)

func AboutPage(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {

		return c.Render(http.StatusOK, "about--about", map[string]interface{}{})
	}
}
