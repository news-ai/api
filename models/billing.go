package models

import (
	"net/http"
	"time"

	"google.golang.org/appengine/log"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type Billing struct {
	Base

	StripeId     string    `json:"-"`
	StripePlanId string    `json:"-"`
	Expires      time.Time `json:"-"`
	HasTrial     bool      `json:"-"`
	IsOnTrial    bool      `json:"-"`
	IsAgency     bool      `json:"-"`
	IsCancel     bool      `json:"-"`

	ReasonForCancel string `json:"-"`

	ReasonNotPurchase  string `json:"-"`
	FeedbackAfterTrial string `json:"-"`

	TrialEmailSent bool `json:"-"`

	CardsOnFile []string `json:"-"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (bi *Billing) Create(c context.Context, r *http.Request, currentUser User) (*Billing, error) {
	bi.CreatedBy = currentUser.Id
	bi.Created = time.Now()
	_, err := bi.Save(c)
	return bi, err
}

/*
* Update methods
 */

// Function to save a new billing into App Engine
func (bi *Billing) Save(c context.Context) (*Billing, error) {
	// Update the Updated time
	bi.Updated = time.Now()

	k, err := nds.Put(c, bi.BaseKey(c, "Billing"), bi)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	bi.Id = k.IntID()
	return bi, nil
}
