package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/log"

	"github.com/qedus/nds"
)

type Invite struct {
	Email        string `json:"email"`
	PersonalNote string `json:"personalnote"`
}

type UserInviteCode struct {
	Base

	InviteCode string `json:"invitecode"`
	Email      string `json:"email"`
	IsUsed     bool   `json:"isused"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (uic *UserInviteCode) Create(c context.Context, r *http.Request, currentUser User) (*UserInviteCode, error) {
	// Create user
	uic.CreatedBy = currentUser.Id
	uic.Created = time.Now()
	uic.IsUsed = false

	_, err := uic.Save(c)
	return uic, err
}

/*
* Update methods
 */

// Function to save a new user into App Engine
func (uic *UserInviteCode) Save(c context.Context) (*UserInviteCode, error) {
	uic.Updated = time.Now()
	k, err := nds.Put(c, uic.BaseKey(c, "UserInviteCode"), uic)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	uic.Id = k.IntID()
	return uic, nil
}

// Function to save a new user into App Engine
func (uic *UserInviteCode) Delete(c context.Context) (*UserInviteCode, error) {
	err := nds.Delete(c, uic.BaseKey(c, "UserInviteCode"))
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	return uic, nil
}
