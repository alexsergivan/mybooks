package userbook

import (
	"html/template"
	"log"
	"time"

	"github.com/alexsergivan/mybooks/services"
	"github.com/labstack/echo/v4"

	"gorm.io/gorm"
)

type BookRating struct {
	BookID    string `gorm:"primaryKey;size:191"`
	Book      Book   `gorm:"references:ID"`
	UserID    int    `gorm:"primaryKey"`
	User      User
	Rate      float64
	Comment   string ""
	CreatedAt time.Time
}

type Book struct {
	ID            string `gorm:"primaryKey"`
	Title         string
	Subtitle      string
	PublishedDate string
	GoogleID      string
	Authors       []*Author `gorm:"many2many:book_authors;"`
	Description   template.HTML
	CategoryName  string   `gorm:"size:191"`
	Category      Category `gorm:"references:Name"`
	Thumbnail     string
	Image         string
	BookRatings   []BookRating `gorm:"foreignKey:BookID"`
	CreatedAt     time.Time
}

type Author struct {
	Name  string  `gorm:"primaryKey"`
	Books []*Book `gorm:"many2many:book_authors;"`
}

type Category struct {
	Name string `gorm:"primaryKey"`
}

type BookRates struct {
	Rate   string
	BookId string
	Count  int
	Book   *Book
}

type BookWithRate struct {
	Rate float64
	Book *Book
}

func GetBookByID(id string, db *gorm.DB) Book {
	var b Book
	db.Where("id = ?", id).Preload("Authors").First(&b)

	return b
}

func GetBookRatings(id string, db *gorm.DB) []*BookRating {
	var ratings []*BookRating
	result := db.Where("book_id = ?", id).Preload("User").Preload("Book").Preload("Book.Authors").Order("created_at desc").Find(&ratings)
	if result.Error != nil {
		log.Println(result.Error)
	}
	return ratings
}

func GetBookRatingsCount(id string, db *gorm.DB) int64 {
	var count int64
	db.Model(&BookRating{}).Where("book_id = ?", id).Count(&count)

	return count
}

func GetAverageBookRating(id string, db *gorm.DB) float64 {
	var avRate float64
	result := db.Table("book_ratings").Select("AVG(rate) as rate").Where("book_id = ?", id).Find(&avRate)
	if result.Error != nil {
		log.Println(result.Error)
	}
	return avRate
}

func GetBestBooks(db *gorm.DB, duration time.Time, limit int) []BookRates {
	var books []BookRates
	result := db.Table("book_ratings").Select("AVG(rate) as rate, book_id, COUNT(rate) as count").Where("created_at > ?", duration).Order("rate desc").Group("book_id").Limit(limit).Find(&books)
	if result.Error != nil {
		log.Println(result.Error)
	}

	return books
}

func GetBestRatedBooks(db *gorm.DB, duration time.Time) []BookRates {
	bookRates := GetBestBooks(db, duration, 20)
	for k, br := range bookRates {
		b := GetBookByID(br.BookId, db)
		bookRates[k].Book = &b
	}

	return bookRates
}

func GetBooksRatings(db *gorm.DB) []*BookRating {
	var ratings []*BookRating
	result := db.Preload("User").Preload("Book").Preload("Book.Authors").Order("created_at desc").Find(&ratings)
	if result.Error != nil {
		log.Println(result.Error)
	}
	return ratings
}

func GetRatingsCount(db *gorm.DB) int64 {
	var count int64
	db.Model(&BookRating{}).Count(&count)

	return count
}

func GetBooks(db *gorm.DB) []*Book {
	var b []*Book

	result := db.Model(&Book{}).Preload("Authors").Order("created_at desc").Find(&b)
	if result.Error != nil {
		log.Println(result.Error)
	}

	return b
}

func GetBooksLight(db *gorm.DB) []*Book {
	var b []*Book

	result := db.Model(&Book{}).Order("created_at desc").Find(&b)
	if result.Error != nil {
		log.Println(result.Error)
	}

	return b
}

func GetBooksWithRating(db *gorm.DB, c echo.Context, pageSize int) []*BookWithRate {
	books := GetBooks(db.Scopes(services.Paginate(c, pageSize)))
	var br []*BookWithRate

	for _, book := range books {
		br = append(br, &BookWithRate{
			Rate: GetAverageBookRating(book.ID, db),
			Book: book,
		})
	}
	return br
}

func GetBooksCount(db *gorm.DB) int64 {
	var count int64
	db.Model(&Book{}).Count(&count)

	return count
}

func GetBestBooksByUser(db *gorm.DB, userID int, duration time.Time, limit int) []BookRates {
	var books []BookRates
	result := db.Table("book_ratings").Select("AVG(rate) as rate, book_id, COUNT(rate) as count").Where("created_at > ? AND user_id = ?", duration, userID).Group("book_id").Limit(limit).Find(&books)
	if result.Error != nil {
		log.Println(result.Error)
	}

	return books
}

func GetBestRatedBooksByUser(db *gorm.DB, userID int, duration time.Time) []BookRates {
	bookRates := GetBestBooksByUser(db, userID, duration, 18)
	for k, br := range bookRates {
		b := GetBookByID(br.BookId, db)
		bookRates[k].Book = &b
	}

	return bookRates
}
