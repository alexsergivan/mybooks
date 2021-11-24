package userbook

import (
	"html/template"
	"log"
	"regexp"
	"strings"
	"time"

	"google.golang.org/api/books/v1"

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

func GetBooksListGroupedByLetter(db *gorm.DB, letter string) map[string][]Book {
	var b []struct {
		ID           string
		Title        string
		Subtitle     string
		CategoryName string
		Letter       string    `json:"letter"`
		CreatedAt    time.Time `json:"created_at"`
	}
	booklist := map[string][]Book{}

	if letter == "all" {
		err := db.Select("id, title, subtitle, category_name, substr(REPLACE(title, '\"', ''), 1, 1) AS letter, created_at").
			Order("title ASC").
			Table("books").
			Find(&b).
			Error
		if err != nil {
			log.Println(err)
		}
	} else {
		err := db.Select("id, title, subtitle, category_name, substr(REPLACE(title, '\"', ''), 1, 1) AS letter, created_at").
			Order("title ASC").
			Table("books").
			Where("substr(REPLACE(title, '\"', ''),1,1)=?", letter).
			Find(&b).
			Error
		if err != nil {
			log.Println(err)
		}
	}

	isAlpha := regexp.MustCompile(`^[A-Za-z]+$`).MatchString
	for _, book := range b {
		if !isAlpha(book.Letter) {
			continue
		}
		booklist[book.Letter] = append(booklist[book.Letter], Book{
			ID:           book.ID,
			Title:        book.Title,
			Subtitle:     book.Subtitle,
			CategoryName: book.CategoryName,
			CreatedAt:    book.CreatedAt,
		})
	}

	return booklist
}

func GetAlphabet(db *gorm.DB) []string {
	var letters []string
	isAlpha := regexp.MustCompile(`^[A-Za-z]+$`).MatchString
	err := db.Select("substr(REPLACE(title, '\"', ''),1,1) as letter").
		Order("letter ASC").
		Table("books").
		Group("letter").
		Find(&letters).
		Error

	if err != nil {
		log.Println(err)
	}

	var finalLetters []string
	for _, letter := range letters {
		if isAlpha(letter) {
			finalLetters = append(finalLetters, letter)
		}
	}

	return finalLetters
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
	bookRates := GetBestBooksByUser(db, userID, duration, 10)
	for k, br := range bookRates {
		b := GetBookByID(br.BookId, db)
		bookRates[k].Book = &b
	}

	return bookRates
}

func ConvertVolumesToBooks(volumes *books.Volumes) []Book {
	var convertedBooks []Book
	for _, volumeItem := range volumes.Items {
		convertedBooks = append(convertedBooks, ConvertVolumeToBook(volumeItem))
	}
	return convertedBooks
}

func ConvertVolumeToBook(volume *books.Volume) Book {
	var authors []*Author
	if len(volume.VolumeInfo.Authors) > 0 {
		for _, author := range volume.VolumeInfo.Authors {
			authors = append(authors, &Author{
				Name: author,
			})
		}
	}

	var category string
	if len(volume.VolumeInfo.Categories) > 0 {
		// Use 1st category.
		category = volume.VolumeInfo.Categories[0]
	} else {
		category = "No Category"
	}

	image, thumbnail := "", ""
	if volume.VolumeInfo.ImageLinks == nil {
		image, thumbnail = "", ""
	} else {
		if volume.VolumeInfo.ImageLinks.Large != "" {
			image = strings.Replace(volume.VolumeInfo.ImageLinks.Large, "http://", "https://", -1)
		} else {
			if volume.VolumeInfo.ImageLinks.Medium != "" {
				image = strings.Replace(volume.VolumeInfo.ImageLinks.Medium, "http://", "https://", -1)
			}
		}

		if volume.VolumeInfo.ImageLinks.Thumbnail != "" {
			thumbnail = strings.Replace(volume.VolumeInfo.ImageLinks.Thumbnail, "http://", "https://", -1)
		}
	}
	return Book{
		ID:            volume.Id,
		Title:         volume.VolumeInfo.Title,
		Subtitle:      volume.VolumeInfo.Subtitle,
		PublishedDate: volume.VolumeInfo.PublishedDate,
		GoogleID:      volume.Id,
		Authors:       authors,
		Category: Category{
			Name: category,
		},
		Description: template.HTML(volume.VolumeInfo.Description),
		Thumbnail:   thumbnail,
		Image:       image,
	}
}
