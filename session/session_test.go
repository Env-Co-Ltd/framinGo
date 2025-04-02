package session

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/alexedwards/scs/v2"
)

func TestSession_InitSession(t *testing.T) {
	c := &Session{
		CookieLifetime: "100",
		CookiePersist:  "true",
		CookieName:     "Framingo",
		CookieDomain:   "localhost",
		SessionType:    "cookie",
	}
  var sm *scs.SessionManager

	ses := c.InitSession()

  var sessKind reflect.Kind
  var sessType reflect.Type

  rv := reflect.ValueOf(ses)

  for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
    fmt.Println("For roop:", rv.Kind(), rv.Type(), rv)
    sessKind = rv.Kind()
    sessType = rv.Type()

    rv = rv.Elem()
  }

  if !rv.IsValid() {
    t.Error("Invalid type or kind; kind:", rv.Kind(), "type:", rv.Type())
  }

  if sessKind != reflect.ValueOf(sm).Kind() {
    t.Error("Kinds do not match:", sessKind, "Expected:", reflect.ValueOf(sm).Kind())
  }

  if sessType != reflect.ValueOf(sm).Type() {
    t.Error("Types do not match:", sessType, "Expected:", reflect.ValueOf(sm).Type())
  }

}

