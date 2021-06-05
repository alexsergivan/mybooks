package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/alexsergivan/mybooks/book"

	"github.com/alexsergivan/mybooks/resolvers"

	"github.com/alexsergivan/mybooks/session"

	"github.com/alexsergivan/mybooks/renderer"

	"github.com/labstack/echo/v4/middleware"

	"github.com/alexsergivan/mybooks/auth"
	"github.com/labstack/echo/v4"

	"github.com/alexsergivan/mybooks/repository"
	"github.com/alexsergivan/mybooks/userbook"
)

//var once sync.Once
//var store *gormstore.Store

//go:embed views/*
var tpls embed.FS

func main() {

	db := repository.GetDB()

	store := session.GetStore()

	e := echo.New()
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 3,
	}))

	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf",
	}))

	e.Static("/", "./public")

	//assetHandler := http.FileServer(getFileSystem())
	//e.GET("/", echo.WrapHandler(assetHandler))
	//e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", assetHandler)))
	booksApiService := book.NewBooksApiService()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))
	e.Renderer = renderer.NewView(tpls)

	e.GET("/", resolvers.HomePage(db, store, booksApiService)).Name = "home"

	e.GET("/about", resolvers.AboutPage(db, store)).Name = "about"
	e.GET("/privacy-policy", resolvers.PrivacyPage(db, store)).Name = "policy"
	e.GET("/sitemap.xml", resolvers.GetSitemap(db, store))
	e.GET("/robots.txt", resolvers.GetRobots())

	userGroup := e.Group("/user", auth.IsAuthMiddleware())
	// Pass store and db, redirect to user/id
	userGroup.GET("", resolvers.ProfilePage(db, store, booksApiService)).Name = "userHome"
	userGroup.GET("/rate-book", resolvers.RateBookForm(db, store)).Name = "rateBook"
	userGroup.POST("/rate-book", resolvers.RateBookSubmit(db, store, booksApiService)).Name = "rateBookSubmit"

	e.GET("/reader/:id", resolvers.ProfilePage(db, store, booksApiService)).Name = "userProfile"

	e.GET("/login", userbook.LoginPage, auth.IsNotAuthMiddleware()).Name = "login"

	authHandler := auth.NewAuthHandler(db, store)

	e.GET("/auth/:provider/callback", authHandler.ProviderCallback())

	e.GET("/auth/:provider", authHandler.StartAuth())
	e.GET("/logout", authHandler.LogOut())

	bookGroup := e.Group("/book")
	bookGroup.GET("/:id", resolvers.BookProfilePage(db, store, booksApiService)).Name = "bookProfile"

	apiGroup := e.Group("/api", auth.IsAuthMiddleware())
	apiGroup.GET("/books/search", book.BooksAutocomplete(booksApiService))

	e.GET("/books", resolvers.BooksPage(db, store)).Name = "books"
	e.GET("/books/search", resolvers.BooksSearchAutocomplete(db, store))

	s := &http.Server{
		Addr:         ":3000",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      http.DefaultServeMux,
	}
	s.SetKeepAlivesEnabled(false)

	e.HTTPErrorHandler = customHTTPErrorHandler

	e.Logger.Fatal(e.StartServer(s))
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	if err := c.Render(code, fmt.Sprintf("errors--%d", code), map[string]interface{}{}); err != nil {
		c.Logger().Error(err)
		_ = c.Render(code, "errors--error", map[string]interface{}{})

	}
	c.Logger().Error(err)
}

//go:embed public
var embededFiles embed.FS

func getFileSystem() http.FileSystem {

	fsys, err := fs.Sub(embededFiles, "public")
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
}
