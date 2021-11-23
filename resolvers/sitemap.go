package resolvers

import (
	"encoding/xml"
	"net/http"
	"strconv"
	"time"

	"github.com/alexsergivan/mybooks/cache"

	"github.com/alexsergivan/mybooks/userbook"
	"github.com/wader/gormstore/v2"
	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
)

type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	Urls    []*Url   `xml:"url"`
}

type Url struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
}

func NewSitemap() *Sitemap {
	return &Sitemap{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		Urls:  make([]*Url, 0),
	}
}

func (s *Sitemap) AddUrl(url *Url) {
	s.Urls = append(s.Urls, url)
}

func GetSitemap(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		sitemap := getSitemapContent(c, db)

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationXMLCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		return xml.NewEncoder(c.Response()).Encode(sitemap)
	}
}

func GetRobots() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "User-agent: *\nAllow: /")
	}
}

func getSitemapContent(c echo.Context, db *gorm.DB) *Sitemap {
	ristrettoCache := cache.NewRistrettoCache()
	cacheKey := "Sitemap"

	sitemap, found := ristrettoCache.Get(cacheKey)
	if !found {
		sitemap := NewSitemap()
		sitemapItems := userbook.GetBooksLight(db)
		usersItems := userbook.GetAllUsers(db)

		for _, sitemapItem := range sitemapItems {
			sitemap.AddUrl(&Url{
				Loc:        "https://" + c.Request().Host + "/book/" + sitemapItem.ID,
				LastMod:    sitemapItem.CreatedAt.Format("2006-01-02"),
				ChangeFreq: "daily",
			})
		}

		for _, user := range usersItems {
			sitemap.AddUrl(&Url{
				Loc:        "https://" + c.Request().Host + "/reader/" + strconv.Itoa(int(user.ID)),
				LastMod:    user.UpdatedAt.Format("2006-01-02"),
				ChangeFreq: "daily",
			})
		}

		sitemap.AddUrl(&Url{
			Loc:        "https://" + c.Request().Host + "/library",
			LastMod:    time.Now().Format("2006-01-02"),
			ChangeFreq: "daily",
		})

		ristrettoCache.Set(cacheKey, sitemap, time.Hour*12)
		time.Sleep(10 * time.Millisecond)

		return sitemap
	}

	return sitemap.(*Sitemap)

}
