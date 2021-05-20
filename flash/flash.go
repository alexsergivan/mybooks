package flash

import (
	"log"

	"github.com/alexsergivan/mybooks/session"
	"github.com/labstack/echo/v4"
)

const flashSessionName = "flash-session"

const MessageTypeMessage = "message"
const MessageTypeError = "error"

func GetMessageTypes() []string {
	return []string{
		MessageTypeError, MessageTypeMessage,
	}
}

func SetFlashMessage(c echo.Context, name string, value string) {
	store := session.GetCookieStore()
	sess, _ := store.Get(c.Request(), flashSessionName)
	sess.AddFlash(value, name)

	err := sess.Save(c.Request(), c.Response())
	if err != nil {
		log.Println(err)
	}
}

func GetFlashMessage(c echo.Context, name string) ([]string, error) {
	store := session.GetCookieStore()
	sess, _ := store.Get(c.Request(), flashSessionName)
	fm := sess.Flashes(name)
	if len(fm) > 0 {
		err := sess.Save(c.Request(), c.Response())
		if err != nil {
			log.Println(err)
		}
		var flashes []string
		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}

		return flashes, nil
	}
	return nil, nil
}
