package resolvers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/alexsergivan/mybooks/cache"

	"github.com/alexsergivan/mybooks/services"

	"github.com/alexsergivan/mybooks/flash"

	"gorm.io/gorm/clause"

	"github.com/alexsergivan/mybooks/book"

	"github.com/alexsergivan/mybooks/userbook"
	"github.com/labstack/echo/v4"
	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"
)

var mutex = sync.RWMutex{}

func RateBookForm(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		mutex.Lock()
		defer mutex.Unlock()

		uID := GetUserIdFromSession(c, storage)

		u := userbook.GetUserByID(int64(*uID), db)
		bookID := c.QueryParam("book")

		var bookRating *userbook.BookRating
		var bookItem *userbook.Book
		if uID != nil && bookID != "" {
			br := userbook.GetUserBookRating(int64(*uID), bookID, db)
			if br.BookID != "" && br.Book.Title != "" {
				bookRating = &br
				bookItem = &bookRating.Book
			} else {
				b := userbook.GetBookByID(bookID, db)
				if b.ID != "" {
					bookItem = &b
				}
			}

		} else {
			bookRating = nil
		}

		if u.Email != nil {
			return c.Render(http.StatusOK, "book--rate", map[string]interface{}{
				"profile":    u,
				"bookRating": bookRating,
				"bookID":     bookID,
				"book":       bookItem,
			})
		}
		return c.Redirect(http.StatusSeeOther, "/")
	}
}

func RateBookSubmit(db *gorm.DB, storage *gormstore.Store, bookApiService *book.BooksApi) echo.HandlerFunc {
	return func(c echo.Context) error {
		bookID := c.FormValue("bookID")
		bookFromApi, err := bookApiService.GetBook(bookID)
		if err != nil {
			flash.SetFlashMessage(c, flash.MessageTypeError, `Please try to select a book again`)
			return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("rateBook"))
		}

		if bookFromApi.ServerResponse.HTTPStatusCode != 200 {
			flash.SetFlashMessage(c, flash.MessageTypeError, fmt.Sprintf(`Something went wrong: %d`, bookFromApi.ServerResponse.HTTPStatusCode))
			return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("rateBook"))
		}

		b := userbook.ConvertVolumeToBook(bookFromApi)

		uID := GetUserIdFromSession(c, storage)
		if uID == nil {
			flash.SetFlashMessage(c, flash.MessageTypeError, `Your session expired. Please Sign In again.`)
			return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("rateBook"))
		}
		rate, _ := strconv.ParseFloat(c.FormValue("rate"), 64)

		if b.Image != "" {
			b.Image, err = services.SaveBookCover(b.Image, bookID, "large")
			if err != nil {
				log.Println(err)
			}
		}
		if b.Thumbnail != "" {
			b.Thumbnail, err = services.SaveBookCover(b.Thumbnail, bookID, "thumbnail")
			if err != nil {
				log.Println(err)
			}
		}

		rating := userbook.BookRating{
			Book: b,
			User: userbook.User{
				Model: gorm.Model{
					ID: *uID,
				},
			},
			Rate:      rate,
			Comment:   c.FormValue("comment"),
			CreatedAt: time.Now(),
		}

		db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&rating)

		if rating.BookID != "" {
			flash.SetFlashMessage(c, flash.MessageTypeMessage, fmt.Sprintf(`Your review of "%s" have been published!`, bookFromApi.VolumeInfo.Title))
		}

		return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("userHome"))
	}
}

func BookProfilePage(db *gorm.DB, storage *gormstore.Store, bookApiService *book.BooksApi) echo.HandlerFunc {
	return func(c echo.Context) error {

		mutex.Lock()
		defer mutex.Unlock()
		id := c.Param("id")
		book := userbook.GetBookByID(id, db)
		if book.Title != "" {
			ristrettoCache := cache.NewRistrettoCache()

			pageSize := 15
			ratingsCount := userbook.GetBookRatingsCount(book.ID, db)

			page, _ := strconv.Atoi(c.QueryParam("page"))
			if page == 0 {
				page++
			}

			var nextPage int
			if page*pageSize < int(ratingsCount) {
				nextPage = page + 1
			}

			params := map[string]interface{}{
				"book":      book,
				"nextPage":  nextPage,
				"rateCount": ratingsCount,
			}
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				params["ratings"] = userbook.GetBookRatings(book.ID, db.Scopes(services.Paginate(c, pageSize)))
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				params["rate"] = userbook.GetAverageBookRating(book.ID, db)
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				avgRating := userbook.GetAverageBookRating(book.ID, db)
				params["stars"] = services.ConvertRateFrom100To5(avgRating)
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				cacheKey := "otherBooks" + book.ID
				otherBooks, found := ristrettoCache.Get(cacheKey)
				if !found {
					bookCat := book.CategoryName
					if bookCat == "No Category" {
						bookCat = "novel"
					}

					oBooks := userbook.ConvertVolumesToBooks(bookApiService.SearchBooksByCategory(bookCat))
					params["otherBooks"] = oBooks
					ristrettoCache.Set(cacheKey, oBooks, time.Minute*15)
					time.Sleep(10 * time.Millisecond)
				} else {
					params["otherBooks"] = otherBooks
				}
			}()
			wg.Wait()

			return c.Render(http.StatusOK, "book--profile", params)
		} else {
			b, err := getBookFromApi(c, id, bookApiService)
			if err != nil {
				flash.SetFlashMessage(c, flash.MessageTypeError, `Something went wrong.`)
				return c.Redirect(http.StatusSeeOther, "/")
			}
			if b.Image != "" {
				b.Image, err = services.SaveBookCover(b.Image, id, "large")
				if err != nil {
					log.Println(err)
				}
			}
			if b.Thumbnail != "" {
				b.Thumbnail, err = services.SaveBookCover(b.Thumbnail, id, "thumbnail")
				if err != nil {
					log.Println(err)
				}
			}

			db.Create(&b)

			return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("bookProfile", id))
		}
	}
}

func BooksPage(db *gorm.DB, storage *gormstore.Store, bookApiService *book.BooksApi) echo.HandlerFunc {
	return func(c echo.Context) error {
		// New books about something.
		if c.QueryParam("new-about") != "" {
			aboutTerm := c.QueryParam("new-about")
			newBooks := userbook.ConvertVolumesToBooks(bookApiService.SearchNewBooks(aboutTerm))

			return c.Render(http.StatusOK, "books--new-books", map[string]interface{}{
				"books":    newBooks,
				"nextPage": nil,
				"topic":    aboutTerm,
			})

		}

		// Best rated books about something (rated and stored in our DB)
		if c.QueryParam("best-about") != "" {

		}

		dbScopes := db.Scopes()
		unescapedCateg := ""
		if c.QueryParam("category") != "" {
			var err error
			unescapedCateg, err = url.QueryUnescape(c.QueryParam("category"))
			if err != nil {
				log.Println(err)
			} else {
				dbScopes = db.Scopes(BookCategory(c, unescapedCateg))
			}
		}

		pageSize := 15
		booksCount := userbook.GetBooksCount(dbScopes)

		page, _ := strconv.Atoi(c.QueryParam("page"))
		if page == 0 {
			page++
		}

		var nextPage int
		if page*pageSize < int(booksCount) {
			nextPage = page + 1
		}

		b := userbook.GetBooksWithRating(db, dbScopes, c, pageSize)

		return c.Render(http.StatusOK, "books--books", map[string]interface{}{
			"books":          b,
			"nextPage":       nextPage,
			"activeCategory": unescapedCateg,
			"category":       c.QueryParam("category"),
		})
	}
}

func BooksSearchAutocomplete(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		books := userbook.GetBooks(db.Scopes(BookTitleLike(c, c.QueryParam("q"))))
		var booksItems []*userbook.Book
		for _, volume := range books {
			if volume != nil {
				booksItems = append(booksItems, volume)
			}
		}

		return c.JSON(http.StatusOK, booksItems)
	}
}

func BookTitleLike(c echo.Context, title string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("title LIKE ?", "%"+title+"%")
	}
}

func BookCategory(c echo.Context, category string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("category_name LIKE ?", "%"+category+"%")
	}
}

func getBookFromApi(c echo.Context, bookID string, bookApiService *book.BooksApi) (userbook.Book, error) {
	bookFromApi, err := bookApiService.GetBook(bookID)
	if err != nil {
		flash.SetFlashMessage(c, flash.MessageTypeError, `Please try to select a book again`)
		return userbook.Book{}, c.Redirect(http.StatusSeeOther, c.Echo().Reverse("home"))
	}

	if bookFromApi.ServerResponse.HTTPStatusCode != 200 {
		flash.SetFlashMessage(c, flash.MessageTypeError, fmt.Sprintf(`Something went wrong: %d`, bookFromApi.ServerResponse.HTTPStatusCode))
		return userbook.Book{}, c.Redirect(http.StatusSeeOther, c.Echo().Reverse("home"))
	}

	return userbook.ConvertVolumeToBook(bookFromApi), nil
}
