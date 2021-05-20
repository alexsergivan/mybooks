package repository

import (
	"log"
	"sync"

	"github.com/alexsergivan/mybooks/config"
	"github.com/alexsergivan/mybooks/userbook"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var once sync.Once

type Repository interface {
	Connect() (*gorm.DB, error)
	migrate()
}

type MySQLRepository struct {
	db     *gorm.DB
	host   string
	port   string
	dbName string
	user   string
	pass   string
}

func NewMySQLRepository(host, port, dbName, user, pass string) *MySQLRepository {
	return &MySQLRepository{
		host:   host,
		port:   port,
		dbName: dbName,
		user:   user,
		pass:   pass,
	}

}

func (r *MySQLRepository) Connect() (*gorm.DB, error) {
	dsn := r.user + ":" + r.pass + "@tcp(" + r.host + ":" + r.port + ")/" + r.dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	r.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	r.migrate()
	return r.db, err
}

func (r *MySQLRepository) migrate() {
	err := r.db.Set("gorm:table_options", "charset=utf8mb4").AutoMigrate(userbook.User{}, userbook.Book{}, userbook.Category{}, userbook.Author{}, userbook.BookRating{})
	if err != nil {
		log.Println(err)
	}
}

func GetDB() *gorm.DB {
	once.Do(func() {
		repo := NewMySQLRepository(
			config.Config("DATABASE_HOST"),
			config.Config("DATABASE_PORT"),
			config.Config("DATABASE_NAME"),
			config.Config("DATABASE_USER"),
			config.Config("DATABASE_PASS"),
		)
		var err error
		db, err = repo.Connect()
		if err != nil {
			log.Println(err)
		}
	})

	return db
}
