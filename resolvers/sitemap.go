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

type SiteMapIndex struct {
	XMLName  xml.Name       `xml:"sitemapindex"`
	Xmlns    string         `xml:"xmlns,attr"`
	Sitemaps []*SitemapItem `xml:"sitemap"`
}

type SitemapItem struct {
	Loc string `xml:"loc"`
}

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

func NewSitemapIndex() SiteMapIndex {
	return SiteMapIndex{
		Xmlns:    "http://www.sitemaps.org/schemas/sitemap/0.9",
		Sitemaps: make([]*SitemapItem, 0),
	}
}

func (si *SiteMapIndex) AddSitemap(sitemap *SitemapItem) {
	si.Sitemaps = append(si.Sitemaps, sitemap)
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

func GetSitemapIndex(db *gorm.DB, storage *gormstore.Store) echo.HandlerFunc {
	return func(c echo.Context) error {
		sitemap := getSitemapIndexContent(c, db)

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

func getSitemapIndexContent(c echo.Context, db *gorm.DB) SiteMapIndex {
	ristrettoCache := cache.NewRistrettoCache()
	cacheKey := "SitemapIndex"
	siteMapIndex, found := ristrettoCache.Get(cacheKey)
	if !found {
		siteMapIndex := NewSitemapIndex()
		alphabet := userbook.GetAlphabet(db)
		for _, letter := range alphabet {
			siteMapIndex.AddSitemap(&SitemapItem{Loc: "https://" + c.Request().Host + "/sitemap/" + letter + "/sitemap.xml"})
		}

		siteMapIndex.AddSitemap(&SitemapItem{Loc: "https://" + c.Request().Host + "/sitemap/users/sitemap.xml"})
		siteMapIndex.AddSitemap(&SitemapItem{Loc: "https://" + c.Request().Host + "/sitemap/library/sitemap.xml"})

		ristrettoCache.Set(cacheKey, siteMapIndex, time.Hour*24)
		time.Sleep(10 * time.Millisecond)
		return siteMapIndex
	}

	return siteMapIndex.(SiteMapIndex)
}

func getSitemapContent(c echo.Context, db *gorm.DB) *Sitemap {
	ristrettoCache := cache.NewRistrettoCache()
	typeVal := c.Param("type")
	cacheKey := "Sitemap_" + typeVal

	sitemap, found := ristrettoCache.Get(cacheKey)
	if !found {
		sitemap := NewSitemap()

		switch typeVal {
		case "users":
			usersItems := userbook.GetAllUsers(db)
			for _, user := range usersItems {
				sitemap.AddUrl(&Url{
					Loc:        "https://" + c.Request().Host + "/reader/" + strconv.Itoa(int(user.ID)),
					LastMod:    user.UpdatedAt.Format("2006-01-02T15:04:05-07:00"),
					ChangeFreq: "daily",
				})
			}
		case "library":
			sitemap.AddUrl(&Url{
				Loc:        "https://" + c.Request().Host + "/library",
				LastMod:    time.Now().Format("2006-01-02T15:04:05-07:00"),
				ChangeFreq: "daily",
			})
			alphabet := userbook.GetAlphabet(db)
			for _, letter := range alphabet {
				sitemap.AddUrl(&Url{
					Loc:        "https://" + c.Request().Host + "/library?letter=" + letter,
					LastMod:    time.Now().Format("2006-01-02T15:04:05-07:00"),
					ChangeFreq: "daily",
				})
			}
		default:
			sitemapItems := userbook.GetBooksListGroupedByLetter(db, typeVal)
			for _, sitemapItem := range sitemapItems {
				for _, book := range sitemapItem {
					sitemap.AddUrl(&Url{
						Loc:        "https://" + c.Request().Host + "/book/" + book.ID,
						LastMod:    book.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
						ChangeFreq: "daily",
					})
				}

			}
		}

		ristrettoCache.Set(cacheKey, sitemap, time.Hour*24)
		time.Sleep(10 * time.Millisecond)

		return sitemap
	}

	return sitemap.(*Sitemap)

}
