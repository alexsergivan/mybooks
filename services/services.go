package services

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/alexsergivan/mybooks/config"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/alexsergivan/mybooks/spaces"

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

	////open a file for writing
	//dir, _ := os.Getwd()
	//
	//file, err := os.Create(dir + "/public" + filePath)
	//if err != nil {
	//	return "", err
	//}
	//defer file.Close()
	//
	//// Use io.Copy to just dump the response body to the file. This supports huge files
	//_, err = io.Copy(file, response.Body)
	//if err != nil {
	//	return "", err
	//}

	buf := bytes.NewBuffer(make([]byte, 0, response.ContentLength))
	_, readErr := buf.ReadFrom(response.Body)
	if readErr != nil {
		return "", readErr
	}
	body := buf.Bytes()

	s3spaces := spaces.GetSpacesClient()
	input := &s3.PutObjectInput{
		Bucket: aws.String(config.Config("BUCKET")),
		Key:    aws.String(filePath),
		ACL:    aws.String("public-read"),
		Body:   bytes.NewReader(body),
		//Metadata: map[string]*string{
		//	"x-amz-meta-my-key": aws.String("your-value"),
		//},
	}
	_, err := s3spaces.PutObject(input)

	if err != nil {
		return "", err
	}

	return "https://mybooks-static-bucket.fra1.cdn.digitaloceanspaces.com" + filePath, nil
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

func ConvertRateFrom100ToEmoji(rate float64) string {
	if rate >= 95 {
		return "🔥"
	}
	if rate < 95 && rate >= 90 {
		return "💖"
	}
	if rate >= 75 && rate < 90 {
		return "🤩"
	}
	if rate > 60 && rate < 75 {
		return "👌"
	}
	if rate >= 40 && rate <= 60 {
		return "😐"
	}

	if rate > 20 && rate < 40 {
		return "👎"
	}

	if rate > 5 && rate <= 20 {
		return "🤢"
	}

	if rate <= 5 {
		return "💩"
	}

	return ""
}
