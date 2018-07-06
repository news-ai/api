package routes

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"

	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/news-ai/api/controllers"

	"github.com/news-ai/web/api"
	nError "github.com/news-ai/web/errors"
)

func handleInvite(c context.Context, r *http.Request, id string) (interface{}, error) {
	return nil, errors.New("method not implemented")
}

func handleInvites(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		val, included, count, total, err := controllers.GetInvites(c, r)
		return api.BaseResponseHandler(val, included, count, total, err, r)
	case "POST":
		return api.BaseSingleResponseHandler(controllers.CreateInvite(c, r))
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the contacts.
func InvitesHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	val, err := handleInvites(c, w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Invite handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /users/<id> route.
func InviteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	val, err := handleInvite(c, r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Invite handling error", err.Error())
	}
	return
}
