package resolvers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/alexsergivan/mybooks/services"

	"github.com/alexsergivan/mybooks/flash"
	"github.com/alexsergivan/mybooks/userbook"
	"github.com/labstack/echo/v4"
	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"
)

const readingQueueSlug = "reading-queue"
const readingQueueName = "Reading Queue"

func BookshelvesPage(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Param("id")
		uID := GetUserIdFromSession(c, storage)
		if nUserID, err := strconv.ParseInt(userID, 10, 64); err == nil {
			bookshelves := userbook.GetUSerBookshelves(db, nUserID)
			return c.Render(http.StatusOK, "bookshelf--user-bookshelves", map[string]interface{}{
				"bookshelves": bookshelves,
				"ownPage":     uID != nil && int64(*uID) == nUserID,
				"userID":      userID,
			})
		}

		return c.Redirect(http.StatusSeeOther, "/books")
	}
}

func BookshelfPage(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Param("id")
		uID := GetUserIdFromSession(c, storage)
		bookshelfSlug := c.Param("bookshelfSlug")
		if userID != "" && bookshelfSlug != "" {
			if nUserID, err := strconv.ParseInt(userID, 10, 64); err == nil {
				bookshelf := userbook.GetUserBookShelfBySlug(db, nUserID, bookshelfSlug)
				templateData := map[string]interface{}{
					"bookShelf": bookshelf,
					"ownPage":   uID != nil && int64(*uID) == nUserID,
					"userID":    userID,
				}

				if c.QueryParam("widget") == "1" {
					if c.QueryParam("dark") == "1" {
						templateData["dark"] = true
					}
					return c.Render(http.StatusOK, "base:widget---widget", templateData)
				}
				return c.Render(http.StatusOK, "bookshelf--user-bookshelf", templateData)
			}

		}
		return c.Redirect(http.StatusSeeOther, "/books")

	}
}

func AddBookToBookshelfPage(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		bookshelfSlug := c.Param("bookshelfSlug")
		bookID := c.Param("bookId")

		uID := GetUserIdFromSession(c, storage)
		u := userbook.GetUserByID(int64(*uID), db)

		bookShelf := userbook.GetUserBookShelfBySlug(db, int64(u.ID), bookshelfSlug)
		if bookShelf.Slug == "" {
			if bookshelfSlug == readingQueueSlug {
				bookShelf = userbook.SaveBookshelf(db, readingQueueName, readingQueueSlug, "", int64(u.ID))
			} else {
				flash.SetFlashMessage(c, flash.MessageTypeError, "This Book Shelf does not exists")
				return c.Redirect(http.StatusSeeOther, "/")
			}
		}

		b := userbook.GetBookByID(bookID, db)
		if b.ID == "" {
			flash.SetFlashMessage(c, flash.MessageTypeError, "This Book does not exist")
			return c.Redirect(http.StatusSeeOther, "/")
		}

		ub := userbook.GetUserBookFromBooksShelf(db, int64(u.ID), bookshelfSlug, bookID)

		if c.QueryParam("delete") != "" {
			return c.Render(http.StatusOK, "bookshelf--delete-book", map[string]interface{}{
				"profile":   u,
				"bookShelf": bookShelf,
				"bookID":    bookID,
				"book":      b,
				"userBook":  ub,
			})
		}

		if ub.BookID != "" {
			return c.Render(http.StatusOK, "bookshelf--edit-book", map[string]interface{}{
				"profile":   u,
				"bookShelf": bookShelf,
				"bookID":    bookID,
				"book":      b,
				"userBook":  ub,
			})
		}

		return c.Render(http.StatusOK, "bookshelf--bookshelf", map[string]interface{}{
			"profile":   u,
			"bookShelf": bookShelf,
			"bookID":    bookID,
			"book":      b,
		})

	}

}

func AddBookToSelectedBookshelfPage(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		bookID := c.Param("bookId")

		uID := GetUserIdFromSession(c, storage)
		u := userbook.GetUserByID(int64(*uID), db)

		b := userbook.GetBookByID(bookID, db)
		if b.ID == "" {
			flash.SetFlashMessage(c, flash.MessageTypeError, "This Book does not exist")
			return c.Redirect(http.StatusSeeOther, "/")
		}

		bookshelves := userbook.GetUSerBookshelves(db, int64(*uID))

		return c.Render(http.StatusOK, "bookshelf--add-to-bookshelf", map[string]interface{}{
			"profile":     u,
			"bookshelves": bookshelves,
			"bookID":      bookID,
			"book":        b,
		})
	}
}

func AddBookshelfPage(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		uID := GetUserIdFromSession(c, storage)
		u := userbook.GetUserByID(int64(*uID), db)
		bookshelfSlug := c.Param("bookshelfSlug")
		bookShelf := &userbook.Bookshelf{}
		if bookshelfSlug != "" {
			bs := userbook.GetUserBookShelfBySlug(db, int64(u.ID), bookshelfSlug)
			bookShelf = &bs
		} else {
			bookShelf = nil
		}
		if c.QueryParam("delete") != "" && bookshelfSlug != readingQueueSlug {
			return c.Render(http.StatusOK, "bookshelf--delete-bookshelf", map[string]interface{}{
				"profile":   u,
				"bookshelf": bookShelf,
			})
		}

		return c.Render(http.StatusOK, "bookshelf--new-bookshelf", map[string]interface{}{
			"profile":   u,
			"bookshelf": bookShelf,
		})
	}
}

func AddBookshelfSubmit(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.FormValue("name")
		description := c.FormValue("description")
		uID := GetUserIdFromSession(c, storage)

		slug := c.FormValue("bookshelfSlug")
		// If it's a new bookshelf.
		if slug == "" {
			slug = services.NormalizeForUrl(name, "en")
		}

		bookshelf := userbook.SaveBookshelf(db, name, slug, description, int64(*uID))

		if bookshelf.Name != "" {
			flash.SetFlashMessage(c, flash.MessageTypeMessage, fmt.Sprintf(`Bookshelf "%s" have been saved. Don't forget to add some books to it 😉 You can do it from any book profile page.`, bookshelf.Name))
			return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("bookshelf", int64(*uID), bookshelf.Slug))
		}

		flash.SetFlashMessage(c, flash.MessageTypeError, "Something went wrong. Please try again.")
		return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("addBookshelf"))
	}
}

func AddBookToBookshelfSubmit(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		uID := GetUserIdFromSession(c, storage)
		bookShelfSlug := c.FormValue("bookShelfSlug")

		bookID := c.FormValue("bookID")

		if bookShelfSlug == "" {
			flash.SetFlashMessage(c, flash.MessageTypeError, "Please select a bookshelf.")
			return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("addBookToSelectedBookshelf", bookID))
		}

		userBookID := c.FormValue("userBookID")
		if userBookID != "" {
			n, err := strconv.ParseInt(userBookID, 10, 64)
			if err == nil {
				log.Print(err)
			}
			ub := userbook.GetUserBookByID(db, n)
			bookStatus := c.FormValue("status")
			if bookStatus == "on" {
				ub.Status = 1
			} else {
				ub.Status = 0
			}

			err = userbook.SaveUserBook(db, ub)
			if err != nil {
				flash.SetFlashMessage(c, flash.MessageTypeError, "Something went wrong. Please try again.")
				return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("addBookToBookshelf", bookShelfSlug, bookID))
			} else {
				flash.SetFlashMessage(c, flash.MessageTypeMessage, fmt.Sprintf(`The book has been updated`))
			}
		} else {

			ub := userbook.GetUserBookFromBooksShelf(db, int64(*uID), bookShelfSlug, bookID)
			if ub.BookID != "" {
				flash.SetFlashMessage(c, flash.MessageTypeError, "This book already exists in your selected bookshelf.")
				return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("bookshelf", int64(*uID), bookShelfSlug))
			}

			bookShelf := userbook.GetUserBookShelfBySlug(db, int64(*uID), bookShelfSlug)

			b := userbook.GetBookByID(bookID, db)
			err := bookShelf.AddBook(db, b)
			if err != nil {
				log.Println(err)
				flash.SetFlashMessage(c, flash.MessageTypeError, "Something went wrong. Please try again.")
				return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("addBookToBookshelf", bookShelfSlug, bookID))
			}
			flash.SetFlashMessage(c, flash.MessageTypeMessage, fmt.Sprintf(`The book "%s" has been added to your "%s" bookshelf`, b.Title, bookShelf.Name))
		}

		return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("bookshelf", int64(*uID), bookShelfSlug))
	}
}

func DeleteBookshelfPage(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		bookshelfSlug := c.FormValue("bookShelfSlug")
		uID := GetUserIdFromSession(c, storage)

		if bookshelfSlug == readingQueueSlug {
			flash.SetFlashMessage(c, flash.MessageTypeError, "It's not possible to remove "+readingQueueName)
			return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("bookshelves", int64(*uID)))
		}
		err := userbook.DeleteBookshelf(db, int64(*uID), bookshelfSlug)
		if err != nil {
			log.Print(err)
			flash.SetFlashMessage(c, flash.MessageTypeError, "Something went wrong. Please try again.")

		} else {
			flash.SetFlashMessage(c, flash.MessageTypeMessage, fmt.Sprintf(`Bookshelf have been deleted`))
		}
		return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("bookshelves", int64(*uID)))

	}
}

func DeleteBookFromBookshelfSubmit(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		userBookID := c.FormValue("userBookID")
		bookshelfID := c.FormValue("bookShelfID")
		bookshelfSlug := c.FormValue("bookShelfSlug")

		uID := GetUserIdFromSession(c, storage)

		if userBookID == "" || bookshelfID == "" {
			return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("bookshelf", int64(*uID), bookshelfSlug))
		}

		userBookIDn, err := strconv.ParseInt(userBookID, 10, 64)
		if err == nil {
			log.Print(err)
		}

		bookshelfIDn, err := strconv.ParseInt(bookshelfID, 10, 64)
		if err == nil {
			log.Print(err)
		}

		err = userbook.DeleteUserBookFromBookshelf(db, userBookIDn, bookshelfIDn)
		if err != nil {
			log.Print(err)
			flash.SetFlashMessage(c, flash.MessageTypeError, "Something went wrong. Please try again.")

		} else {
			flash.SetFlashMessage(c, flash.MessageTypeMessage, fmt.Sprintf(`The book has been deleted from your bookshelf`))
		}
		return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("bookshelf", int64(*uID), bookshelfSlug))

	}
}
