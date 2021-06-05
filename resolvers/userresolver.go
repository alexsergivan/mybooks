package resolvers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/alexsergivan/mybooks/book"

	"github.com/alexsergivan/mybooks/services"

	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"

	"github.com/gorilla/sessions"

	"github.com/alexsergivan/mybooks/repository"

	"github.com/alexsergivan/mybooks/session"
	"github.com/alexsergivan/mybooks/userbook"
	"github.com/labstack/echo/v4"
)

func GetCurrentUser(c echo.Context) *userbook.User {
	email := GetUserEmailFromSession(c)
	if email != nil {
		u := userbook.Get(&userbook.User{Email: email}, repository.GetDB())
		return &u
	}

	return nil
}

func GetUserEmailFromSession(c echo.Context) *string {
	sess := GetUserSession(session.GetStore(), c.Request())

	userVal := sess.Values["user"]
	email, ok := userVal.(string)
	if !ok {
		return nil
	}

	return &email
}

func GetUserIdFromSession(c echo.Context, store sessions.Store) *uint {
	sess := GetUserSession(store, c.Request())

	userVal := sess.Values["userID"]
	id, ok := userVal.(uint)
	if !ok {
		return nil
	}

	return &id
}

func GetUserSession(store sessions.Store, req *http.Request) *sessions.Session {
	sess, err := store.Get(req, "session")
	if err != nil {
		log.Println(err)
	}

	return sess
}

func ProfilePage(db *gorm.DB, storage *gormstore.Store, booksApi *book.BooksApi) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			uID := GetUserIdFromSession(c, storage)
			return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("userProfile", int64(*uID)))
		}
		nId, err := strconv.Atoi(id)
		if err != nil {

			return c.Redirect(http.StatusSeeOther, "/")
		}
		uID := GetUserIdFromSession(c, storage)

		u := userbook.GetUserByID(int64(nId), db)

		pageSize := 15
		ratingsCount := userbook.GetUserBookRatingsCount(int64(nId), db)

		page, _ := strconv.Atoi(c.QueryParam("page"))
		if page == 0 {
			page++
		}

		var nextPage int
		if page*pageSize < int(ratingsCount) {
			nextPage = page + 1
		}

		if u.Email != nil {
			return c.Render(http.StatusOK, "user--profile", map[string]interface{}{
				"profile":              u,
				"ratings":              userbook.GetUserBookRatings(int64(nId), db.Scopes(services.Paginate(c, pageSize))),
				"topRatings":           userbook.GetTopUserBookRatings(int64(nId), db, 10),
				"page":                 page,
				"nextPage":             nextPage,
				"ratingsCount":         ratingsCount,
				"ownPage":              uID != nil && int(*uID) == nId,
				"avRate":               int(userbook.GetAverageRatingByUser(int64(nId), db)),
				"positiveRatingsCount": userbook.GePositiveBookRatingsFromUserCount(int64(nId), db),
				"negativeRatingsCount": userbook.GeNegativeBookRatingsFromUserCount(int64(nId), db),
				//TODO: move it to the homepage and add logic when go yo book page, create a new page, if does not exist.
				//"recommendedBooks": userbook.GetBookRecommendations(nId, db, booksApi),
			})
		}
		return c.Redirect(http.StatusSeeOther, "/")
	}
}
