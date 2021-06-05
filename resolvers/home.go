package resolvers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/alexsergivan/mybooks/book"

	"github.com/alexsergivan/mybooks/cache"

	"github.com/alexsergivan/mybooks/services"

	"github.com/alexsergivan/mybooks/userbook"
	"github.com/labstack/echo/v4"
	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"
)

func HomePage(db *gorm.DB, storage *gormstore.Store, booksApi *book.BooksApi) echo.HandlerFunc {
	return func(c echo.Context) error {
		ristrettoCache := cache.NewRistrettoCache()
		user := GetCurrentUser(c)
		uKey := 0
		if user != nil {
			uKey = int(user.ID)
		}
		cacheKey := "HomePage" + c.QueryParam("page") + string(rune(uKey))

		home, found := ristrettoCache.Get(cacheKey)
		if !found {
			home := getHomePageData(c, db, booksApi, uKey)
			ristrettoCache.Set(cacheKey, home, time.Minute*15)
			time.Sleep(10 * time.Millisecond)
			return c.Render(http.StatusOK, "index--index", home)
		}

		return c.Render(http.StatusOK, "index--index", home.(map[string]interface{}))
	}
}

func getHomePageData(c echo.Context, db *gorm.DB, booksApi *book.BooksApi, userId int) map[string]interface{} {
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
		"topBooksDay":      userbook.GetBestRatedBooks(db, time.Now().AddDate(0, 0, -1)),
		"topBooksMonth":    userbook.GetBestRatedBooks(db, time.Now().AddDate(0, -1, 0)),
		"topBooksYear":     userbook.GetBestRatedBooks(db, time.Now().AddDate(-1, 0, 0)),
		"nextPage":         nextPage,
		"ratings":          userbook.GetBooksRatings(db.Scopes(services.Paginate(c, pageSize))),
		"recommendedBooks": userbook.GetBookRecommendations(userId, db, booksApi),
	}
	return r
}
