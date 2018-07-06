package controllers

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/qedus/nds"

	"github.com/news-ai/api/models"
)

func GetUserBilling(c context.Context, r *http.Request, user models.User) (models.Billing, error) {
	if user.BillingId == 0 {
		return models.Billing{}, errors.New("No billing for this user")
	}
	// Get the billing by id
	var billing models.Billing
	billingId := datastore.NewKey(c, "Billing", "", user.BillingId, nil)
	err := nds.Get(c, billingId, &billing)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Billing{}, err
	}

	billing.Format(billingId, "billings")

	return billing, nil
}
