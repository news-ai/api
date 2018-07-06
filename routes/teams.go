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

func handleTeamActions(c context.Context, r *http.Request, id string, action string) (interface{}, error) {
	return nil, errors.New("method not implemented")
}

func handleTeam(c context.Context, r *http.Request, id string) (interface{}, error) {
	switch r.Method {
	case "GET":
		return api.BaseSingleResponseHandler(controllers.GetTeam(c, id))
	}
	return nil, errors.New("method not implemented")
}

func handleTeams(c context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	switch r.Method {
	case "GET":
		val, included, count, total, err := controllers.GetTeams(c, r)
		return api.BaseResponseHandler(val, included, count, total, err, r)
	case "POST":
		return api.BaseSingleResponseHandler(controllers.CreateTeam(c, r))
	}
	return nil, errors.New("method not implemented")
}

// Handler for when the user wants all the agencies.
func TeamsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	val, err := handleTeams(c, w, r)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Team handling error", err.Error())
	}
	return
}

// Handler for when there is a key present after /users/<id> route.
func TeamHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	val, err := handleTeam(c, r, id)

	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Team handling error", err.Error())
	}
	return
}

// Handler for when the user wants to perform an action on the publications
func TeamActionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	c := appengine.NewContext(r)
	id := ps.ByName("id")
	action := ps.ByName("action")

	val, err := handleTeamActions(c, r, id, action)
	if err == nil {
		err = ffjson.NewEncoder(w).Encode(val)
	}

	if err != nil {
		nError.ReturnError(w, http.StatusInternalServerError, "Team handling error", err.Error())
	}
	return
}
