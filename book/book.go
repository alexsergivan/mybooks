package book

// check user package
//type Book struct {
//	ID            string `gorm:"primaryKey"`
//	Title         string
//	Subtitle      string
//	PublishedDate string
//	GoogleID      string
//	Authors       []*Author `gorm:"many2many:book_authors;"`
//	Description   template.HTML
//	CategoryName  string   `gorm:"size:191"`
//	Category      Category `gorm:"references:Name"`
//	Thumbnail     string
//	Image         string
//}
//
//type Author struct {
//	Name  string  `gorm:"primaryKey"`
//	Books []*Book `gorm:"many2many:book_authors;"`
//}
//
//type Category struct {
//	Name string `gorm:"primaryKey"`
//}
//
//func GetBookByID(id string, db *gorm.DB) Book {
//	var b Book
//	db.Where("id = ?", id).Preload("Authors").First(&b)
//
//	return b
//}
