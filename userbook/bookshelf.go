package userbook

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const toReadBookshelfName = "toread"

type Bookshelf struct {
	gorm.Model
	Name        string
	Description string
	Slug        string
	User        User
	UserBooks   []UserBook `gorm:"many2many:bookshelf_user_books;"`
	UserID      int64
}

func (bs *Bookshelf) AddBook(db *gorm.DB, book Book) error {
	bs.UserBooks = append(bs.UserBooks, UserBook{
		User:   bs.User,
		Book:   book,
		UserID: bs.UserID,
		BookID: book.ID,
		Status: 0,
	})

	//db.Clauses(clause.OnConflict{
	//	UpdateAll: true,
	//}).Create(bs)

	if err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(bs).Error; err != nil {
		return err
	}
	return nil
}

func SaveBookshelf(db *gorm.DB, name, slug, description string, userID int64) Bookshelf {
	shelf := Bookshelf{}
	bs := GetUserBookShelfBySlug(db, userID, slug)
	if bs.Slug != "" {
		shelf = bs
	}

	shelf.Name = name
	shelf.Slug = slug
	shelf.UserID = userID
	shelf.Description = description

	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&shelf)

	return shelf
}

func IsUserBooksShelfExists(db *gorm.DB, userID int64, slug string) bool {
	var bookshelf Bookshelf
	db.Where("user_id= ?", userID).Where("slug=?", slug).First(&bookshelf)
	return bookshelf.Slug != ""
}

func GetUserBookShelfBySlug(db *gorm.DB, userID int64, slug string) Bookshelf {
	var bookshelf Bookshelf
	db.Preload("User").Preload("UserBooks", func(db *gorm.DB) *gorm.DB {
		return db.Order("status desc, updated_at desc")
	}).Preload("UserBooks.Book").Where("user_id= ?", userID).Where("slug=?", slug).Order("updated_at desc").First(&bookshelf)

	return bookshelf
}

func GetBookShelfByID(db *gorm.DB, id int64) Bookshelf {
	var bookshelf Bookshelf
	db.Where("id= ?", id).Preload("User").Preload("UserBooks").First(&bookshelf)

	return bookshelf
}

func GetUSerBookshelves(db *gorm.DB, userID int64) []Bookshelf {
	var bookshelves []Bookshelf
	db.Where("user_id=?", userID).Preload("UserBooks", func(db *gorm.DB) *gorm.DB {
		return db.Order("status desc, updated_at desc")
	}).Preload("UserBooks.Book").Preload("User").Order("updated_at desc").Find(&bookshelves)

	return bookshelves
}

func AddBookToBookshelf(db *gorm.DB, userID int64, name string, book Book) {
	bookshelf := GetUserBookShelfBySlug(db, userID, name)
	bookshelf.AddBook(db, book)
}

func GetUserBookFromBooksShelf(db *gorm.DB, userID int64, bookshelfSlug, bookID string) UserBook {

	res := UserBook{}

	db.Model(&UserBook{}).Joins("inner join bookshelf_user_books on bookshelf_user_books.user_book_id = user_books.id").Joins("inner join bookshelves on bookshelves.id = bookshelf_user_books.bookshelf_id and bookshelves.slug=?", bookshelfSlug).Where("user_books.book_id=?", bookID).Where("user_books.user_id=?", userID).Scan(&res)

	return res
}

func DeleteBookshelf(db *gorm.DB, userID int64, slug string) error {
	res := db.Where("user_id=? AND slug=?", userID, slug).Delete(&Bookshelf{})
	if res.Error != nil {
		return res.Error
	}
	return nil
}
