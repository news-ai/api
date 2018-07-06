package models

import (
	"net/http"
	"time"

	"google.golang.org/appengine/log"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type Agency struct {
	Base

	Name  string `json:"name"`
	Email string `json:"email"`

	BillingId int64 `json:"billingid"`

	Administrators []int64 `json:"administrators" datastore:",noindex" apiModel:"User"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (a *Agency) Create(c context.Context, r *http.Request, currentUser User) (*Agency, error) {
	a.CreatedBy = currentUser.Id
	a.Created = time.Now()
	_, err := a.Save(c)
	return a, err
}

/*
* Update methods
 */

// Function to save a new agency into App Engine
func (a *Agency) Save(c context.Context) (*Agency, error) {
	// Update the Updated time
	a.Updated = time.Now()

	k, err := nds.Put(c, a.BaseKey(c, "Agency"), a)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	a.Id = k.IntID()
	return a, nil
}

/*
* Action methods
 */

func (a *Agency) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := SetField(a, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
