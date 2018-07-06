package billing

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"

	"github.com/news-ai/api/models"
	"github.com/news-ai/tabulae/emails"
)

func AddFreeTrialToUser(r *http.Request, user models.User, plan string) (int64, error) {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	sc := client.New(os.Getenv("STRIPE_SECRET_KEY"), stripe.NewBackends(httpClient))

	// https://stripe.com/docs/api
	// Create new customer in Stripe
	params := &stripe.CustomerParams{
		Email:    user.Email,
		Plan:     plan + "-trial",
		Quantity: uint64(1),
	}

	customer, err := sc.Customers.New(params)
	if err != nil {
		log.Errorf(c, "%v", err)
		return 0, err
	}

	_, billingId, err := user.SetStripeId(c, r, user, customer.ID, plan, true, true)
	if err != nil {
		log.Errorf(c, "%v", err)
		return billingId, err
	}
	return billingId, nil
}

func AddPlanToUser(r *http.Request, user models.User, userBilling *models.Billing, plan string, duration string, coupon string, originalPlan string) error {
	c := appengine.NewContext(r)
	httpClient := urlfetch.Client(c)
	sc := client.New(os.Getenv("STRIPE_SECRET_KEY"), stripe.NewBackends(httpClient))

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

	// Only considers plans currently that moving from trial. Not changing plans.
	// Cancel all past subscriptions they had
	for i := 0; i < len(customer.Subs.Values); i++ {
		sc.Subs.Cancel(customer.Subs.Values[i].ID, nil)
	}

	// Start a new subscription without trial (they already went through the trial)
	params := &stripe.SubParams{
		Customer: customer.ID,
		Plan:     plan,
	}

	if duration == "annually" {
		params.Plan = plan + "-yearly"
	}

	if coupon != "" {
		coupon = strings.ToUpper(coupon)
		params.Coupon = coupon
	}

	if strings.ToLower(coupon) == "favorites" && duration == "annually" {
		return errors.New("Sorry - you can't use this coupon code on a yearly plan. Please switch the monthly one to use this!")
	}

	if strings.ToLower(coupon) == "prcouture" && duration == "annually" {
		return errors.New("Sorry - you can't use this coupon code on a yearly plan. Please switch the monthly one to use this!")
	}

	if strings.ToLower(coupon) == "curious" && duration == "annually" {
		return errors.New("Sorry - you can't use this coupon code on a yearly plan. Please switch the monthly one to use this!")
	}

	if strings.ToLower(coupon) == "prconsultants" && duration == "annually" {
		return errors.New("Sorry - you can't use this coupon code on a yearly plan. Please switch the monthly one to use this!")
	}

	newSub, err := sc.Subs.New(params)
	if err != nil {
		var stripeError StripeError
		err = json.Unmarshal([]byte(err.Error()), &stripeError)
		if err != nil {
			log.Errorf(c, "%v", err)
			return errors.New("We had an error setting your subscription")
		}

		log.Errorf(c, "%v", err)
		return errors.New(stripeError.Message)
	}

	// Return if there are any errors
	expiresAt := time.Unix(newSub.PeriodEnd, 0)
	userBilling.Expires = expiresAt
	userBilling.StripePlanId = plan
	userBilling.IsOnTrial = false
	userBilling.Save(c)

	// Set the user to be an active being on the platform again
	user.IsActive = true
	user.Save(c)

	currentPrice := PlanAndDurationToPrice(originalPlan, duration)
	billAmount := "$" + fmt.Sprintf("%0.2f", currentPrice)
	paidAmount := "$" + fmt.Sprintf("%0.2f", currentPrice)

	ExpiresAt := expiresAt.Format("2006-01-02")

	emailDuration := "a monthly"
	if duration == "annually" {
		emailDuration = "an annual"
	}

	// Email confirmation
	err = emails.AddUserToTabulaePremiumList(c, user, originalPlan, emailDuration, ExpiresAt, billAmount, paidAmount)
	if err != nil {
		log.Errorf(c, "%v", err)
	}

	return nil
}
