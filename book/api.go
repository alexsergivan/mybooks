package book

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"

	"github.com/alexsergivan/mybooks/config"
	"google.golang.org/api/option"

	"google.golang.org/api/books/v1"
)

var once sync.Once
var booksApiService *books.Service

type BooksApi struct {
	svc *books.Service
}

func GetBooksApiService() *books.Service {
	once.Do(func() {
		opts := []option.ClientOption{
			option.WithAPIKey(config.Config("GOOGLE_API_KEY")),
		}
		var err error
		booksApiService, err = books.NewService(context.Background(), opts...)
		if err != nil {
			log.Println(err)
		}
	})

	return booksApiService
}

func NewBooksApiService() *BooksApi {
	return &BooksApi{
		GetBooksApiService(),
	}
}

func (api *BooksApi) SearchBooks(q string) *books.Volumes {
	volumes, err := api.svc.Volumes.List(q).Do()
	if err != nil {
		log.Println(err)
	}
	return volumes
}

func (api *BooksApi) GetBook(id string) (*books.Volume, error) {
	return api.svc.Volumes.Get(id).Do()
}

func BooksAutocomplete(booksApi *BooksApi) echo.HandlerFunc {
	return func(c echo.Context) error {
		volumes := booksApi.SearchBooks(c.QueryParam("q"))
		var booksItems []*AutocompleteBookItem
		for _, volume := range volumes.Items {
			if volume != nil {
				booksItems = append(booksItems, ConvertFromVolumeToAutocompleteItem(volume))
			}
		}

		return c.JSON(http.StatusOK, booksItems)
	}
}

type AutocompleteBookItem struct {
	Title      string
	Subtitle   string
	GoogleID   string
	Authors    []string
	Categories []string
	Thumbnail  string
}

func ConvertFromVolumeToAutocompleteItem(volume *books.Volume) *AutocompleteBookItem {

	thumb := ""
	if volume.VolumeInfo.ImageLinks != nil {
		thumb = volume.VolumeInfo.ImageLinks.SmallThumbnail
	}
	return &AutocompleteBookItem{
		Title:      volume.VolumeInfo.Title,
		Subtitle:   volume.VolumeInfo.Subtitle,
		GoogleID:   volume.Id,
		Authors:    volume.VolumeInfo.Authors,
		Categories: volume.VolumeInfo.Categories,
		Thumbnail:  thumb,
	}

}
