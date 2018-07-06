package models

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/log"

	"github.com/qedus/nds"
)

type Team struct {
	Base

	Name string `json:"name"`

	AgencyId int64 `json:"agencyid" apiModel:"Agency"`

	MaxMembers int `json:"maxmembers" apiModel:"User"`

	Members []int64 `json:"members" apiModel:"User"`
	Admins  []int64 `json:"admins" apiModel:"User"`
}

/*
* Public methods
 */

/*
* Create methods
 */

// Function to create a new team into App Engine
func (t *Team) Create(c context.Context, r *http.Request, currentUser User) (*Team, error) {
	t.CreatedBy = currentUser.Id
	t.Created = time.Now()

	_, err := t.Save(c)
	return t, err
}

/*
* Update methods
 */

// Function to save a new team into App Engine
func (t *Team) Save(c context.Context) (*Team, error) {
	// Update the Updated time
	t.Updated = time.Now()

	// Save the object
	k, err := nds.Put(c, t.BaseKey(c, "Team"), t)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	t.Id = k.IntID()
	return t, nil
}
