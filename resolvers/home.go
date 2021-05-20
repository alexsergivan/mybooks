package resolvers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/alexsergivan/mybooks/services"

	"github.com/alexsergivan/mybooks/userbook"
	"github.com/labstack/echo/v4"
	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"
)

func HomePage(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {

		pageSize := 6
		ratingsCount := userbook.GetRatingsCount(db)

		page, _ := strconv.Atoi(c.QueryParam("page"))
		if page == 0 {
			page++
		}

		var nextPage int
		if page*pageSize < int(ratingsCount) {
			nextPage = page + 1
		}

		r := map[string]interface{}{
			"topBooksDay":   userbook.GetBestRatedBooks(db, time.Now().AddDate(0, 0, -1)),
			"topBooksMonth": userbook.GetBestRatedBooks(db, time.Now().AddDate(0, -1, 0)),
			"topBooksYear":  userbook.GetBestRatedBooks(db, time.Now().AddDate(-1, 0, 0)),
			"nextPage":      nextPage,
			"ratings":       userbook.GetBooksRatings(db.Scopes(services.Paginate(c, pageSize))),
		}
		return c.Render(http.StatusOK, "index--index", r)
	}
}
