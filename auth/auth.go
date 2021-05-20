package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/alexsergivan/mybooks/services"

	"github.com/alexsergivan/mybooks/resolvers"

	"github.com/labstack/echo/v4"

	"github.com/alexsergivan/mybooks/userbook"

	"gorm.io/gorm"

	"github.com/gorilla/sessions"

	"github.com/alexsergivan/mybooks/config"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type AuthenticatorHandler struct {
	DB    *gorm.DB
	Store sessions.Store
}

func NewAuthHandler(db *gorm.DB, store sessions.Store) *AuthenticatorHandler {
	initAuth(store)
	return &AuthenticatorHandler{
		DB:    db,
		Store: store,
	}
}

func initAuth(store sessions.Store) {
	gothic.Store = store
	goth.UseProviders(
		google.New(config.Config("GOOGLE_CLIENT_ID"), config.Config("GOOGLE_CLIENT_SECRET"), config.Config("GOOGLE_CALLBACK_URL"), "email", "profile"),
	)
}

func (h *AuthenticatorHandler) ProviderCallback() echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := h.completeUserAuth(c.Response(), c.Request(), c.Param("provider"))
		if err == nil {
			return c.Redirect(http.StatusSeeOther, "/user")
		}
		return err
	}
}

func (h *AuthenticatorHandler) StartAuth() echo.HandlerFunc {
	return func(c echo.Context) error {
		u, err := h.completeUserAuth(c.Response(), c.Request(), c.Param("provider"))

		if u == nil || err != nil {
			h.beginUserAuth(c.Response(), c.Request(), c.Param("provider"))
		}

		return nil
	}
}

func (h *AuthenticatorHandler) LogOut() echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := h.Store.Get(c.Request(), "session")
		if err != nil {
			log.Println(err)
		}
		session.Options.MaxAge = -1
		err = h.Store.Save(c.Request(), c.Response(), session)
		if err != nil {
			log.Println(err)
		}

		return c.Redirect(http.StatusSeeOther, "/")
	}
}

func (h *AuthenticatorHandler) completeUserAuth(w http.ResponseWriter, req *http.Request, provider string) (*goth.User, error) {
	ctx := context.WithValue(req.Context(), "provider", provider)
	if u, err := gothic.CompleteUserAuth(w, req.WithContext(ctx)); err == nil {
		h.syncGothUser(u, w, req)
		return &u, nil
	} else {
		return nil, err
	}
}

func (h *AuthenticatorHandler) beginUserAuth(w http.ResponseWriter, req *http.Request, provider string) {
	ctx := context.WithValue(req.Context(), "provider", provider)
	gothic.BeginAuthHandler(w, req.WithContext(ctx))
}

// syncGothUser adds user to the session and db.
func (h *AuthenticatorHandler) syncGothUser(gothUser goth.User, w http.ResponseWriter, req *http.Request) {
	if gothUser.AvatarURL == "" {
		gothUser.AvatarURL = "/images/user.svg"
	} else {
		ava := strings.Replace(gothUser.AvatarURL, "s96-c", "s500-c", -1)
		avatarPath, err := services.SaveUserProfileImage(ava, gothUser.UserID)
		if err != nil {
			log.Println(err)
			gothUser.AvatarURL = ava
		} else {
			gothUser.AvatarURL = avatarPath
		}
	}
	u := &userbook.User{
		Name:         gothUser.Name,
		Email:        &gothUser.Email,
		AvatarURL:    gothUser.AvatarURL,
		GoogleUserID: gothUser.UserID,
	}
	userFromDb := userbook.Get(u, h.DB)
	var userID uint
	if userFromDb.Email == nil {
		userbook.Create(u, h.DB)
		userID = u.ID
	} else {
		userID = userFromDb.ID
		//TEST it!

		if u.AvatarURL != "" && userFromDb.AvatarURL != u.AvatarURL {
			userFromDb.AvatarURL = u.AvatarURL
		}
		if u.Name != "" && userFromDb.Name != u.Name {
			userFromDb.Name = u.Name
		}
		userbook.Update(&userFromDb, h.DB)
		// End test it.
	}

	session := resolvers.GetUserSession(h.Store, req)
	session.Values["user"] = u.Email
	session.Values["userID"] = userID
	err := h.Store.Save(req, w, session)
	if err != nil {
		log.Println(err)
	}
}

func IsAuthenticated(c echo.Context) (bool, string) {
	emailFromSession := resolvers.GetUserEmailFromSession(c)
	if emailFromSession != nil {
		return true, *emailFromSession
	}
	return false, ""
}

func IsAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			isAuth, _ := IsAuthenticated(c)
			if isAuth {
				return next(c)
			}
			return c.Redirect(http.StatusSeeOther, c.Echo().Reverse("login"))
		}
	}
}

func IsNotAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			isAuth, _ := IsAuthenticated(c)
			if !isAuth {
				return next(c)
			}
			return c.Redirect(http.StatusSeeOther, "/")

		}
	}
}
