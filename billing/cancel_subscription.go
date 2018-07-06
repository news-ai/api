package billing

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"

	"github.com/news-ai/api/models"
)

func CancelPlanOfUser(r *http.Request, user models.User, userBilling *models.Billing) error {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	sc := client.New(os.Getenv("STRIPE_SECRET_KEY"), stripe.NewBackends(httpClient))

	if userBilling.IsOnTrial {
		return errors.New("Can not cancel a trial")
	}

	customer, err := sc.Customers.Get(userBilling.StripeId, nil)
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

	// Cancel all plans they might have (they should only have one)
	for i := 0; i < len(customer.Subs.Values); i++ {
		sc.Subs.Cancel(customer.Subs.Values[i].ID, nil)
	}

	userBilling.IsCancel = true
	userBilling.Save(c)

	// Send an email to the user saying that the package will be canceled. Their account will be inactive on
	// their "Expires" date.

	return nil
}
