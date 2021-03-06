package userbook

import (
	"database/sql"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/alexsergivan/mybooks/cache"

	"github.com/alexsergivan/mybooks/book"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name         string
	Email        *string
	AvatarURL    string
	BookRatings  []BookRating `gorm:"foreignKey:UserID"`
	GoogleUserID string
	BookShelves  []Bookshelf `gorm:"foreignKey:UserID"`
}

type UserBook struct {
	gorm.Model
	User   User
	Book   Book
	UserID int64
	BookID string `gorm:"size:191"`
	Status int
}

func LoginPage(c echo.Context) error {

	return c.Render(http.StatusOK, "user--login", map[string]interface{}{})
}

// Get gets user from db.
func Get(user *User, db *gorm.DB) User {
	var u User
	db.First(&u, "email = ?", user.Email)
	return u
}

func GetUserByID(id int64, db *gorm.DB) User {
	var u User
	db.Where("id = ?", id).First(&u, id)

	return u
}

func GetFullUserByID(id int64, db *gorm.DB) User {
	var u User
	db.Where("id = ?", id).Preload("BookRatings.Book").Preload("BookRatings").First(&u, id)

	return u
}

func Create(user *User, db *gorm.DB) {
	db.Create(user)
}

func Update(user *User, db *gorm.DB) {
	db.Save(user)
}

func GetUserBookRatings(id int64, db *gorm.DB) []*BookRating {
	var ratings []*BookRating
	result := db.Where("user_id = ?", id).Preload("Book").Preload("Book.Authors").Order("created_at desc").Find(&ratings)
	if result.Error != nil {
		log.Println(result.Error)
	}
	return ratings
}

func GetUserBookRating(userID int64, bookID string, db *gorm.DB) BookRating {
	var rating BookRating
	if err := db.Where("user_id = ? and book_id = ?", userID, bookID).Preload("Book").Preload("Book.Authors").First(&rating).Error; err != nil {
		return BookRating{}
	}

	return rating
}

func GetUserBookRatingsCount(id int64, db *gorm.DB) int64 {
	var count int64
	db.Model(&BookRating{}).Where("user_id = ?", id).Count(&count)

	return count
}

func GetAverageRatingByUser(id int64, db *gorm.DB) float64 {
	var avRate sql.NullFloat64
	result := db.Table("book_ratings").Select("AVG(rate) as rate").Where("user_id = ?", id).Find(&avRate)
	if result.Error != nil {
		log.Println(result.Error)
	}
	if !avRate.Valid {
		return 0
	}
	return avRate.Float64
}

func GePositiveBookRatingsFromUserCount(id int64, db *gorm.DB) int64 {
	var count int64
	db.Table("book_ratings").Where("user_id = ?", id).Where("rate > 55").Count(&count)

	return count
}

func GeNegativeBookRatingsFromUserCount(id int64, db *gorm.DB) int64 {
	var count int64
	db.Table("book_ratings").Where("user_id = ?", id).Where("rate <= 55").Count(&count)

	return count
}

func GetTopUserBookRatings(id int64, db *gorm.DB, topAmount int) []*BookRating {
	var ratings []*BookRating
	result := db.Where("user_id = ?", id).Limit(topAmount).Preload("Book").Preload("Book.Authors").Order("rate desc").Find(&ratings)
	if result.Error != nil {
		log.Println(result.Error)
	}
	return ratings
}

func GetAllUsers(db *gorm.DB) []*User {
	var u []*User

	result := db.Model(&User{}).Find(&u)
	if result.Error != nil {
		log.Println(result.Error)
	}

	return u
}

func GetBookRecommendations(userID int, db *gorm.DB, booksApi *book.BooksApi) []Book {
	if userID == 0 {
		return nil
	}
	ristrettoCache := cache.NewRistrettoCache()
	cacheKey := "bookRecom" + string(rune(userID))
	recommendations, found := ristrettoCache.Get(cacheKey)
	if !found {

		topBooks := GetBestRatedBooksByUser(db, userID, time.Now().AddDate(0, -1, 0))
		var booksPool []Book

		var wg sync.WaitGroup

		for _, bookItem := range topBooks {
			if bookItem.Book.CategoryName == "No Category" {
				continue
			}
			wg.Add(1)
			bookItem := bookItem
			go func() {
				defer wg.Done()
				booksPool = append(booksPool, ConvertVolumesToBooks(booksApi.SearchBooksByCategory(bookItem.Book.CategoryName))...)
			}()

		}
		wg.Wait()

		if len(booksPool) > 0 {
			var recommendations []Book
			rand.Seed(time.Now().Unix())
			for i := 0; i < 10; i++ {
				randBookKey := rand.Intn(len(booksPool))
				if booksPool[randBookKey].ID != "" {
					recommendations = append(recommendations, booksPool[randBookKey])
					booksPool[randBookKey] = Book{}
				} else {
					i--
				}
			}

			ristrettoCache.Set(cacheKey, recommendations, time.Minute*60)
			time.Sleep(10 * time.Millisecond)

			return recommendations
		} else {
			return booksPool
		}
	}

	return recommendations.([]Book)

}

func GetUserBookByID(db *gorm.DB, userBookID int64) UserBook {
	userBook := UserBook{}
	db.First(&userBook, userBookID)

	return userBook
}

func SaveUserBook(db *gorm.DB, userBook UserBook) error {
	res := db.Save(&userBook)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func DeleteUserBookFromBookshelf(db *gorm.DB, userBookID, bookshelfID int64) error {

	res := db.Exec("DELETE FROM bookshelf_user_books WHERE bookshelf_id=? AND user_book_id=?", bookshelfID, userBookID)
	if res.Error != nil {
		return res.Error
	}
	return nil
}
