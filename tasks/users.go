package tasks

import (
	"net/http"
	"strconv"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/news-ai/api/controllers"

	"github.com/news-ai/tabulae/sync"

	"github.com/news-ai/web/errors"
	"github.com/news-ai/web/utilities"
)

func RefreshUserLiveTokens(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// Get users who's live tokens are going to expire in the next 10 minutes
	users, err := controllers.GetUsersUnauthorized(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		errors.ReturnError(w, http.StatusInternalServerError, "Could not get users", err.Error())
		return
	}

	userIds := []int64{}
	for i := 0; i < len(users); i++ {
		if users[i].LiveAccessTokenExpire.Before(time.Now()) {
			randomString := strconv.FormatInt(users[i].Id, 10)
			randomString = randomString + utilities.RandToken()
			users[i].LiveAccessToken = randomString
			users[i].LiveAccessTokenExpire = time.Now().Local().Add(time.Hour*time.Duration(6) +
				time.Minute*time.Duration(0) +
				time.Second*time.Duration(0))
			users[i].Save(c)
		}

		userIds = append(userIds, users[i].Id)
	}

	sync.UserResourceBulkSync(r, userIds)
}

func MakeUsersInactive(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	users, err := controllers.GetUsersUnauthorized(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		errors.ReturnError(w, http.StatusInternalServerError, "Could not get users", err.Error())
		return
	}

	for i := 0; i < len(users); i++ {
		billing, err := controllers.GetUserBilling(c, r, users[i])
		if err != nil {
			log.Errorf(c, "%v", users[i])
			log.Errorf(c, "%v", err)
			continue
		}

		// For now only consider when they are on trial
		if billing.IsOnTrial {
			if billing.Expires.Before(time.Now()) {
				users[i].IsActive = false
				users[i].Save(c)
				sync.ResourceSync(r, users[i].Id, "User", "create")

				billing.IsOnTrial = false
				billing.IsCancel = true
				billing.Save(c)
			}
		}
	}
}
