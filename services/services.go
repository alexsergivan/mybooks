package services

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"

	"gorm.io/gorm"
)

func SaveBookCover(url, bookID, coverType string) (string, error) {
	return SaveFile(url, "/images/books/"+bookID+"_"+coverType+".jpg")
}

func SaveUserProfileImage(url, userID string) (string, error) {
	return SaveFile(url, "/images/users/"+userID+".jpg")
}

func SaveFile(url, filePath string) (string, error) {
	response, e := http.Get(url)
	if e != nil {
		return "", e
	}
	defer response.Body.Close()

	//open a file for writing
	dir, _ := os.Getwd()

	file, err := os.Create(dir + "/public" + filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func Paginate(c echo.Context, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		page, _ := strconv.Atoi(c.QueryParam("page"))
		if page == 0 {
			page = 1
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func ConvertRateFrom100To5(rate float64) int {
	stars := 1
	if rate >= 90 {
		stars = 5
	}
	if rate >= 75 && rate < 90 {
		stars = 4
	}

	if rate >= 50 && rate < 75 {
		stars = 3
	}

	if rate >= 20 && rate < 50 {
		stars = 2
	}

	return stars
}
