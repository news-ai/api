package billing

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"

	"github.com/news-ai/api/models"
)

func SwitchUserPlanPreview(r *http.Request, user models.User, userBilling *models.Billing, duration, newPlan string) (int64, error) {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	sc := client.New(os.Getenv("STRIPE_SECRET_KEY"), stripe.NewBackends(httpClient))

	customer, err := sc.Customers.Get(userBilling.StripeId, nil)
	if err != nil {
		var stripeError StripeError
		err = json.Unmarshal([]byte(err.Error()), &stripeError)
		if err != nil {
			log.Errorf(c, "%v", err)
			return 0.0, errors.New("We had an error getting your user")
		}

		log.Errorf(c, "%v", err)
		return 0.0, errors.New(stripeError.Message)
	}

	if duration == "annually" {
		newPlan = newPlan + "-yearly"
	}

	if customer.Subs.Count > 0 {
		prorationDate := time.Now().Unix()

		invoiceParams := &stripe.InvoiceParams{
			Customer:         customer.ID,
			Sub:              customer.Subs.Values[0].ID,
			SubPlan:          newPlan,
			SubProrationDate: prorationDate,
		}
		invoice, err := sc.Invoices.GetNext(invoiceParams)

		if err != nil {
			log.Errorf(c, "%v", err)
			return 0.0, err
		}

		var cost int64 = 0
		for _, invoiceItem := range invoice.Lines.Values {
			if invoiceItem.Period.Start == prorationDate {
				cost += invoiceItem.Amount
			}
		}

		return cost, nil
	}

	return 0.00, nil
}

func SwitchUserPlan(r *http.Request, user models.User, userBilling *models.Billing, newPlan string) error {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	sc := client.New(os.Getenv("STRIPE_SECRET_KEY"), stripe.NewBackends(httpClient))

	_, err := sc.Customers.Get(userBilling.StripeId, nil)
	if err != nil {
		var stripeError StripeError
		err = json.Unmarshal([]byte(err.Error()), &stripeError)
		if err != nil {
			log.Errorf(c, "%v", err)
			return errors.New("We had an error getting your user")
		}

		log.Errorf(c, "%v", err)
		return errors.New(stripeError.Message)
	}

	return nil
}
