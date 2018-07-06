package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/julienschmidt/httprouter"

	apiControllers "github.com/news-ai/api/controllers"

	"github.com/news-ai/oauth2/outlook"
	"github.com/news-ai/web/utilities"

	"golang.org/x/oauth2"
)

type OutlookResponse struct {
	OdataContext string `json:"@odata.context"`
	OdataID      string `json:"@odata.id"`
	ID           string `json:"Id"`
	EmailAddress string `json:"EmailAddress"`
	DisplayName  string `json:"DisplayName"`
	Alias        string `json:"Alias"`
	MailboxGUID  string `json:"MailboxGuid"`
}

var (
	outlookOauthConfig = &oauth2.Config{
		RedirectURL:  "https://tabulae.newsai.org/api/auth/outlookcallback",
		ClientID:     os.Getenv("OUTLOOKAUTHKEY"),
		ClientSecret: os.Getenv("OUTLOOKAUTHSECRET"),
		Scopes: []string{
			"https://outlook.office.com/mail.readwrite",
			"https://outlook.office.com/mail.send",
			"openid",
			"profile",
			"offline_access",
		},
		Endpoint: outlook.Endpoint,
	}
)

// Handler to redirect user to the Outlook OAuth2 page
func OutlookLoginHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	// Make sure the user has been logged in when at outlook auth
	user, err := apiControllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		fmt.Fprintln(w, "user not logged in")
		return
	}

	// Generate a random state that we identify the user with
	state := utilities.RandToken()

	// Save the session for each of the users
	session, err := Store.Get(r, "sess")
	if err != nil {
		log.Errorf(c, "%v", err)
	}

	session.Values["state"] = state
	session.Values["outlook"] = "yes"
	session.Values["outlook_email"] = user.Email

	if r.URL.Query().Get("next") != "" {
		session.Values["next"] = r.URL.Query().Get("next")
	}

	err = session.Save(r, w)
	if err != nil {
		log.Errorf(c, "%v", err)
	}

	// Redirect the user to the login page
	url := outlookOauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, 302)
}

// Handler to redirect user to the Google OAuth2 page
func RemoveOutlookHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	// Make sure the user has been logged in when at oulook auth
	user, err := apiControllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		fmt.Fprintln(w, "user not logged in")
		return
	}

	user.Outlook = false
	apiControllers.SaveUser(c, r, &user)

	if r.URL.Query().Get("next") != "" {
		returnURL := r.URL.Query().Get("next")
		if err != nil {
			http.Redirect(w, r, returnURL, 302)
			return
		}
	}

	http.Redirect(w, r, "https://tabulae.newsai.co/settings", 302)
	return
}

func OutlookCallbackHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)
	session, err := Store.Get(r, "sess")
	if err != nil {
		log.Infof(c, "%v", err)
		fmt.Fprintln(w, "aborted")
		return
	}

	if r.URL.Query().Get("state") != session.Values["state"] {
		log.Errorf(c, "%v", "no state match; possible csrf OR cookies not enabled")
		fmt.Fprintln(w, "no state match; possible csrf OR cookies not enabled")
		return
	}

	tkn, err := outlookOauthConfig.Exchange(c, r.URL.Query().Get("code"))
	if err != nil {
		log.Errorf(c, "%v", "there was an issue getting your token")
		fmt.Fprintln(w, "there was an issue getting your token")
		return
	}

	if !tkn.Valid() {
		log.Errorf(c, "%v", "retreived invalid token")
		fmt.Fprintln(w, "retreived invalid token")
		return
	}

	// Make sure the user has been logged in when at outlook auth
	user, err := apiControllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		fmt.Fprintln(w, "user not logged in")
		return
	}

	client := urlfetch.Client(c)

	req, _ := http.NewRequest("GET", "https://outlook.office.com/api/v2.0/me", nil)
	req.Header.Add("Authorization", "Bearer "+tkn.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", "there was an issue getting your token "+err.Error())
		fmt.Fprintln(w, "there was an issue getting your token")
		return
	}
	defer resp.Body.Close()

	// Decode JSON from Google
	decoder := json.NewDecoder(resp.Body)
	var outlookUser OutlookResponse
	err = decoder.Decode(&outlookUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		fmt.Fprintln(w, err.Error())
		return
	}

	user.OutlookEmail = outlookUser.EmailAddress
	user.OutlookAccessToken = tkn.AccessToken
	user.OutlookExpiresIn = tkn.Expiry
	user.OutlookRefreshToken = tkn.RefreshToken
	user.OutlookTokenType = tkn.TokenType

	user.Outlook = true
	user.Gmail = false
	user.ExternalEmail = false

	apiControllers.SaveUser(c, r, &user)

	returnURL := session.Values["next"].(string)
	u, err := url.Parse(returnURL)
	if err != nil {
		http.Redirect(w, r, returnURL, 302)
		return
	}

	http.Redirect(w, r, u.String(), 302)
	return
}
