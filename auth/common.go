package auth

import (
	"errors"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"

	"github.com/news-ai/gaesessions"

	"github.com/news-ai/api/utils"
)

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
	Hd            string `json:"hd"`
}

var Store = gaesessions.NewMemcacheDatastoreStore("", "",
	gaesessions.DefaultNonPersistentSessionDuration,
	[]byte(os.Getenv("SECRETKEY")))

func SetRedirectURL() {
	googleOauthConfig.RedirectURL = utils.APIURL + "/auth/googlecallback"
}

// Gets the email of the current user that is logged in
func GetCurrentUserEmail(r *http.Request) (string, error) {
	session, err := Store.Get(r, "sess")
	if err != nil {
		return "", errors.New("No user logged in")
	}

	if session.Values["email"] == nil {
		return "", errors.New("No user logged in")
	}

	return session.Values["email"].(string), nil
}

// Gets the email of the current user that is logged in
func GetCurrentUserId(r *http.Request) (int64, error) {
	session, err := Store.Get(r, "sess")
	if err != nil {
		return 0, errors.New("No user logged in")
	}

	if session.Values["id"] == nil {
		return 0, errors.New("No user logged in")
	}

	return session.Values["id"].(int64), nil
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := Store.Get(r, "sess")
	delete(session.Values, "state")
	delete(session.Values, "id")
	delete(session.Values, "email")
	session.Save(r, w)

	if r.URL.Query().Get("next") != "" {
		http.Redirect(w, r, r.URL.Query().Get("next"), 302)
		return
	}

	http.Redirect(w, r, "/api/auth", 302)
}
