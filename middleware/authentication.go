package middleware

import (
	"net/http"
	"strings"

	"google.golang.org/appengine"

	"github.com/news-ai/api/auth"
	apiControllers "github.com/news-ai/api/controllers"
	"github.com/news-ai/api/utils"

	"github.com/news-ai/web/errors"
)

func UpdateOrCreateUser(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Basic authentication
	apiKey, _, _ := r.BasicAuth()
	apiKeyValid := false
	if apiKey != "" {
		apiKeyValid = auth.BasicAuthLogin(w, r, apiKey)
	}

	c := appengine.NewContext(r)
	email, err := auth.GetCurrentUserEmail(r)
	if err != nil && !strings.Contains(r.URL.Path, "/api/auth") && !strings.Contains(r.URL.Path, "/static") && !apiKeyValid {
		w.Header().Set("Content-Type", "application/json")
		errors.ReturnError(w, http.StatusUnauthorized, "Authentication Required", "Please login "+utils.APIURL+"/auth/google")
		return
	} else {
		if email != "" {
			apiControllers.AddUserToContext(c, r, email)
		}
	}

	next(w, r)
}
