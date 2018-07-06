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
	apiModels "github.com/news-ai/api/models"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/emails"

	"github.com/news-ai/web/utilities"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "https://tabulae.newsai.org/api/auth/googlecallback",
		ClientID:     os.Getenv("GOOGLEAUTHKEY"),
		ClientSecret: os.Getenv("GOOGLEAUTHSECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

	gmailOauthConfig = &oauth2.Config{
		RedirectURL:  "https://tabulae.newsai.org/api/auth/googlecallback",
		ClientID:     os.Getenv("GOOGLEAUTHKEY"),
		ClientSecret: os.Getenv("GOOGLEAUTHSECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/gmail.readonly",
			"https://www.googleapis.com/auth/gmail.compose",
			"https://www.googleapis.com/auth/gmail.send",
		},
		Endpoint: google.Endpoint,
	}
)

// Handler to redirect user to the Google OAuth2 page
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)
	// Generate a random state that we identify the user with
	state := utilities.RandToken()

	// Save the session for each of the users
	session, err := Store.Get(r, "sess")
	if err != nil {
		log.Errorf(c, "%v", err)
	}

	session.Values["state"] = state
	session.Values["gmail"] = "no"
	session.Values["gmail_email"] = ""

	if r.URL.Query().Get("next") != "" {
		session.Values["next"] = r.URL.Query().Get("next")
	}

	err = session.Save(r, w)
	if err != nil {
		log.Errorf(c, "%v", err)
	}

	// Redirect the user to the login page
	url := googleOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, 302)
	return
}

// Handler to redirect user to the Google OAuth2 page
func RemoveGmailHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	// Make sure the user has been logged in when at gmail auth
	user, err := apiControllers.GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		fmt.Fprintln(w, "user not logged in")
		return
	}

	user.Gmail = false
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

// Handler to redirect user to the Google OAuth2 page
func GmailLoginHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := appengine.NewContext(r)

	// Make sure the user has been logged in when at gmail auth
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
	session.Values["gmail"] = "yes"
	session.Values["gmail_email"] = user.Email

	if r.URL.Query().Get("next") != "" {
		session.Values["next"] = r.URL.Query().Get("next")
	}

	err = session.Save(r, w)
	if err != nil {
		log.Errorf(c, "%v", err)
	}

	// Redirect the user to the login page
	url := gmailOauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, 302)
}

// Handler to get information when callback comes back from Google
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	tkn, err := googleOauthConfig.Exchange(c, r.URL.Query().Get("code"))

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

	client := urlfetch.Client(c)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo?alt=json&access_token=" + tkn.AccessToken)
	if err != nil {
		log.Errorf(c, "%v", err)
		fmt.Fprintln(w, err.Error())
		return
	}
	defer resp.Body.Close()

	// Decode JSON from Google
	decoder := json.NewDecoder(resp.Body)
	var googleUser User
	err = decoder.Decode(&googleUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		fmt.Fprintln(w, err.Error())
		return
	}

	newUser := apiModels.User{}
	newUser.Email = googleUser.Email
	newUser.GoogleId = googleUser.ID
	newUser.FirstName = googleUser.GivenName
	newUser.LastName = googleUser.FamilyName
	newUser.EmailConfirmed = true
	newUser.IsActive = false

	newUser.TokenType = tkn.TokenType
	newUser.GoogleExpiresIn = tkn.Expiry
	newUser.RefreshToken = tkn.RefreshToken
	newUser.AccessToken = tkn.AccessToken
	newUser.GoogleCode = r.URL.Query().Get("code")
	if session.Values["gmail"] == "yes" {
		newUser.Gmail = true
		newUser.Outlook = false
		newUser.ExternalEmail = false

		if session.Values["gmail_email"].(string) != googleUser.Email {
			log.Errorf(c, "%v", "Tried to login with email "+googleUser.Email+" for user "+session.Values["gmail_email"].(string))
			http.Redirect(w, r, "https://tabulae.newsai.co/settings", 302)
			return
		}
	}

	user, _, _ := controllers.RegisterUser(r, newUser)

	session.Values["email"] = googleUser.Email
	session.Values["id"] = newUser.Id
	session.Save(r, w)

	if user.IsActive {
		if session.Values["next"] != nil {
			returnURL := session.Values["next"].(string)
			u, err := url.Parse(returnURL)
			if err != nil {
				http.Redirect(w, r, returnURL, 302)
				return
			}

			if user.LastLoggedIn.IsZero() {
				q := u.Query()
				q.Set("firstTimeUser", "true")
				u.RawQuery = q.Encode()

				err = emails.AddUserToTabulaeTrialList(c, user)
				if err != nil {
					// Redirect user back to login page
					log.Errorf(c, "%v", "Welcome email was not sent for "+user.Email)
					log.Errorf(c, "%v", err)
				}

				user.ConfirmLoggedIn(c)
			}
			http.Redirect(w, r, u.String(), 302)
			return
		}
	} else {
		if user.LastLoggedIn.IsZero() {
			err = emails.AddUserToTabulaeTrialList(c, user)
			if err != nil {
				// Redirect user back to login page
				log.Errorf(c, "%v", "Welcome email was not sent for "+user.Email)
				log.Errorf(c, "%v", err)
			}
		}
		http.Redirect(w, r, "/api/billing/plans/trial", 302)
		return
	}

	http.Redirect(w, r, "/", 302)
	return
}
