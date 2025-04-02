package framinGo

import (
	"net/http"
	"strconv"

	"github.com/justinas/nosurf"
)

func (f *FraminGo) SessionLoad(next http.Handler) http.Handler {
	return f.Session.LoadAndSave(next)
}

func (f *FraminGo) NoSurf(next http.Handler) http.Handler {
  csrfHandler := nosurf.New(next)
  secure, _ :=strconv.ParseBool(f.config.cookie.secure)

  csrfHandler.ExemptGlob("/api/*")

  csrfHandler.SetBaseCookie(http.Cookie{
    HttpOnly: true,
    Path: "/",
    Secure: secure,
    SameSite: http.SameSiteLaxMode,
    Domain: f.config.cookie.domain,
  })
  return csrfHandler
}
