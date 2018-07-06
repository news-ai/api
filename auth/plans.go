package auth

import (
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	apiControllers "github.com/news-ai/api/controllers"

	"github.com/news-ai/api/billing"

	"github.com/gorilla/csrf"
	"github.com/pquerna/ffjson/ffjson"

	nError "github.com/news-ai/web/errors"
)

func TrialPlanPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		user, err := apiControllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			http.Redirect(w, r, "https://tabulae.newsai.co/", 302)
			return
		}

		if !user.IsActive {
			userBilling, err := apiControllers.GetUserBilling(c, r, user)

			// If the user has a user billing
			if err == nil {
				if userBilling.HasTrial && !userBilling.Expires.IsZero() {
					// If the user has already had a trial and has expired
					// This means: If userBilling Expire date is before the current time
					if userBilling.Expires.Before(time.Now()) {
						http.Redirect(w, r, "/api/billing", 302)
						return
					}

					// If the user has already had a trial but it has not expired
					if userBilling.Expires.After(time.Now()) {
						http.Redirect(w, r, "https://tabulae.newsai.co/", 302)
						return
					}
				}
			}

			billingId, err := billing.AddFreeTrialToUser(r, user, "free")
			user.IsActive = true
			user.BillingId = billingId
			user.Save(c)

			// If there was an error creating this person's trial
			if err != nil {
				log.Errorf(c, "%v", err)
				http.Redirect(w, r, "/api/billing/plans/trial", 302)
				return
			}

			// If they have a coupon they want to use (to expire later)
			if user.PromoCode != "" {
				if user.PromoCode == "PRCOUTURE" || user.PromoCode == "GOPUBLIX" {
					userBilling, err := apiControllers.GetUserBilling(c, r, user)
					if err == nil {
						userBilling.Expires = userBilling.Expires.AddDate(0, 3, 0)
						userBilling.Save(c)
					} else {
						log.Infof(c, "%v", err)
					}
				}
			}

			// If not then their is now probably successful so we redirect them back
			returnURL := "https://tabulae.newsai.co/"
			session, _ := Store.Get(r, "sess")
			if session.Values["next"] != nil {
				returnURL = session.Values["next"].(string)
			}
			u, err := url.Parse(returnURL)

			// If there's an error in parsing the return value
			// then returning it.
			if err != nil {
				log.Errorf(c, "%v", err)
				http.Redirect(w, r, returnURL, 302)
				return
			}

			// This would be a bug since they should not be here if they
			// are a firstTimeUser. But we'll allow it to help make
			// experience normal.
			if user.LastLoggedIn.IsZero() {
				q := u.Query()
				q.Set("firstTimeUser", "true")
				u.RawQuery = q.Encode()
				user.ConfirmLoggedIn(c)
			}

			http.Redirect(w, r, u.String(), 302)
			return
		} else {
			// If the user is active then they don't need to start a free trial
			http.Redirect(w, r, "https://tabulae.newsai.co/", 302)
			return
		}
	}
}

func CancelPlanPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		user, err := apiControllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			http.Redirect(w, r, "https://tabulae.newsai.co/", 302)
			return
		}

		userBilling, err := apiControllers.GetUserBilling(c, r, user)

		// If the user has a billing profile
		if err == nil {
			switch userBilling.StripePlanId {
			case "personal":
				userBilling.StripePlanId = "Personal"
			case "consultant":
				userBilling.StripePlanId = "Consultant"
			case "business":
				userBilling.StripePlanId = "Business"
			case "growing":
				userBilling.StripePlanId = "Growing Business"
			}

			userNotActiveNonTrialPlan := true
			if user.IsActive && !userBilling.IsOnTrial {
				userNotActiveNonTrialPlan = false
			}

			data := map[string]interface{}{
				"userNotActiveNonTrialPlan": userNotActiveNonTrialPlan,
				"currentUserPlan":           userBilling.StripePlanId,
				"userEmail":                 user.Email,
				csrf.TemplateTag:            csrf.TemplateField(r),
			}

			t := template.New("cancel.html")
			t, _ = t.ParseFiles("billing/cancel.html")
			t.Execute(w, data)
		} else {
			// If the user does not have billing profile that means that they
			// have not started their trial yet.
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}
	}
}

func CancelPlanHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)

		// To check if there is a user logged in
		user, err := apiControllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				log.Errorf(c, "%v", err)
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			log.Errorf(c, "%v", err)
			http.Redirect(w, r, "https://tabulae.newsai.co/", 302)
			return
		}

		userBilling, err := apiControllers.GetUserBilling(c, r, user)

		// If the user has a billing profile
		if err == nil {
			plan := ""
			switch userBilling.StripePlanId {
			case "personal":
				plan = "Personal"
			case "consultant":
				plan = "Consultant"
			case "business":
				plan = "Business"
			case "growing":
				plan = "Growing Business"
			}

			data := map[string]interface{}{
				"plan":           plan,
				"userEmail":      user.Email,
				csrf.TemplateTag: csrf.TemplateField(r),
			}

			t := template.New("cancelled.html")
			t, _ = t.ParseFiles("billing/confirmation.html")
			t.Execute(w, data)
		} else {
			// If the user does not have billing profile that means that they
			// have not started their trial yet.
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}
	}
}

func ChoosePlanPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		user, err := apiControllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			http.Redirect(w, r, "https://tabulae.newsai.co/", 302)
			return
		}

		userBilling, err := apiControllers.GetUserBilling(c, r, user)

		// If the user has a billing profile
		if err == nil {
			userBilling.StripePlanId = billing.BillingIdToPlanName(userBilling.StripePlanId)

			userNotActiveNonTrialPlan := true
			if user.IsActive && !userBilling.IsOnTrial {
				userNotActiveNonTrialPlan = false
			}

			data := map[string]interface{}{
				"userNotActiveNonTrialPlan": userNotActiveNonTrialPlan,
				"currentUserPlan":           userBilling.StripePlanId,
				"userEmail":                 user.Email,
			}

			t := template.New("plans.html")
			t, _ = t.ParseFiles("billing/plans.html")
			t.Execute(w, data)
		} else {
			// If the user does not have billing profile that means that they
			// have not started their trial yet.
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}
	}
}

func ChooseSwitchPlanHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		plan := r.FormValue("plan")
		duration := r.FormValue("duration")

		// To check if there is a user logged in
		user, err := apiControllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				log.Errorf(c, "%v", err)
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			log.Errorf(c, "%v", err)
			http.Redirect(w, r, "https://tabulae.newsai.co/", 302)
			return
		}

		userBilling, err := apiControllers.GetUserBilling(c, r, user)

		// If the user has a billing profile
		if err == nil {
			originalPlan := plan
			switch plan {
			case "personal":
				plan = "Personal"
			case "consultant":
				plan = "Consultant"
			case "business":
				plan = "Business"
			case "growing":
				plan = "Growing Business"
			}

			missingCard := true
			if len(userBilling.CardsOnFile) > 0 {
				missingCard = false
			}

			price := billing.PlanAndDurationToPrice(plan, duration)
			cost, _ := billing.SwitchUserPlanPreview(r, user, &userBilling, duration, originalPlan)

			data := map[string]interface{}{
				"missingCard": missingCard,
				"price":       price,
				"plan":        plan,
				"duration":    duration,
				"userEmail":   user.Email,
				"difference":  cost,
			}

			t := template.New("switch-confirmation.html")
			t, _ = t.ParseFiles("billing/switch-confirmation.html")
			t.Execute(w, data)
		} else {
			// If the user does not have billing profile that means that they
			// have not started their trial yet.
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}
	}
}

func ChoosePlanHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		plan := r.FormValue("plan")
		duration := r.FormValue("duration")

		// To check if there is a user logged in
		user, err := apiControllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				log.Errorf(c, "%v", err)
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			log.Errorf(c, "%v", err)
			http.Redirect(w, r, "https://tabulae.newsai.co/", 302)
			return
		}

		userBilling, err := apiControllers.GetUserBilling(c, r, user)

		// If the user has a billing profile
		if err == nil {
			switch plan {
			case "personal":
				plan = "Personal"
			case "consultant":
				plan = "Consultant"
			case "business":
				plan = "Business"
			case "growing":
				plan = "Growing Business"
			}

			missingCard := true
			if len(userBilling.CardsOnFile) > 0 {
				missingCard = false
			}

			price := billing.PlanAndDurationToPrice(plan, duration)

			data := map[string]interface{}{
				"missingCard": missingCard,
				"price":       price,
				"plan":        plan,
				"duration":    duration,
				"userEmail":   user.Email,
			}

			t := template.New("confirmation.html")
			t, _ = t.ParseFiles("billing/confirmation.html")
			t.Execute(w, data)
		} else {
			// If the user does not have billing profile that means that they
			// have not started their trial yet.
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}
	}
}

func CheckCouponValid() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		c := appengine.NewContext(r)
		coupon := r.FormValue("coupon")
		duration := r.FormValue("duration")

		if coupon == "" {
			nError.ReturnError(w, http.StatusInternalServerError, "Coupon error", "Please enter a coupon")
			return
		}

		coupon = strings.ToUpper(coupon)

		if coupon == "FAVORITES" && duration == "annually" {
			nError.ReturnError(w, http.StatusInternalServerError, "Coupon error", "Sorry - you can't use this coupon code on a yearly plan. Please switch the monthly one to use this!")
			return
		}

		if coupon == "PRCOUTURE" && duration == "annually" {
			nError.ReturnError(w, http.StatusInternalServerError, "Coupon error", "Sorry - you can't use this coupon code on a yearly plan. Please switch the monthly one to use this!")
			return
		}

		if coupon == "CURIOUS" && duration == "annually" {
			nError.ReturnError(w, http.StatusInternalServerError, "Coupon error", "Sorry - you can't use this coupon code on a yearly plan. Please switch the monthly one to use this!")
			return
		}

		if coupon == "PRCONSULTANTS" && duration == "annually" {
			nError.ReturnError(w, http.StatusInternalServerError, "Coupon error", "Sorry - you can't use this coupon code on a yearly plan. Please switch the monthly one to use this!")
			return
		}

		percentageOff, err := billing.GetCoupon(r, coupon)

		if err == nil {
			val := struct {
				PercentageOff uint64
			}{
				PercentageOff: percentageOff,
			}
			err = ffjson.NewEncoder(w).Encode(val)
		}

		if err != nil {
			log.Errorf(c, "%v", err)
			nError.ReturnError(w, http.StatusInternalServerError, "Coupon error", err.Error())
		}

		return
	}
}

func ConfirmPlanHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		plan := r.FormValue("plan")
		duration := r.FormValue("duration")
		coupon := r.FormValue("coupon")

		// To check if there is a user logged in
		user, err := apiControllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				log.Errorf(c, "%v", err)
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			log.Errorf(c, "%v", err)
			http.Redirect(w, r, "https://tabulae.newsai.co/", 302)
			return
		}

		userBilling, err := apiControllers.GetUserBilling(c, r, user)

		// If the user has a billing profile
		if err == nil {
			originalPlan := plan
			switch plan {
			case "Personal":
				plan = "personal"
			case "Consultant":
				plan = "consultant"
			case "Business":
				plan = "business"
			case "Growing Business":
				plan = "growing"
			}

			err = billing.AddPlanToUser(r, user, &userBilling, plan, duration, coupon, originalPlan)
			hasError := false
			errorMessage := ""
			if err != nil {
				hasError = true
				// Return error to the "confirmation" page
				errorMessage = err.Error()
				log.Errorf(c, "%v", err)
			}

			data := map[string]interface{}{
				"plan":         originalPlan,
				"duration":     duration,
				"hasError":     hasError,
				"errorMessage": errorMessage,
				"userEmail":    user.Email,
			}

			t := template.New("receipt.html")
			t, _ = t.ParseFiles("billing/receipt.html")
			t.Execute(w, data)
		} else {
			// If the user does not have billing profile that means that they
			// have not started their trial yet.
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}
	}
}

func BillingPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		user, err := apiControllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			http.Redirect(w, r, "https://tabulae.newsai.co/", 302)
			return
		}

		userBilling, err := apiControllers.GetUserBilling(c, r, user)

		// If the user has a billing profile
		if err == nil {
			switch userBilling.StripePlanId {
			case "bronze", "personal":
				userBilling.StripePlanId = "Personal"
			case "aluminum", "consultant":
				userBilling.StripePlanId = "Consultant"
			case "silver-1", "silver", "business":
				userBilling.StripePlanId = "Business"
			case "gold-1", "gold", "growing":
				userBilling.StripePlanId = "Growing Business"
			}

			customerBalance, _ := billing.GetCustomerBalance(r, user, &userBilling)
			userPlanExpires := userBilling.Expires.AddDate(0, 0, -1).Format("2006-01-02")

			userbillingHistory, _ := billing.GetCustomerBillingHistory(r, user, &userBilling)
			log.Infof(c, "%v", userbillingHistory)

			data := map[string]interface{}{
				"userBillingPlanExpires": userPlanExpires,
				"userBilling":            userBilling,
				"userEmail":              user.Email,
				"userActive":             user.IsActive,
				"userBalance":            customerBalance,
				"userbillingHistory":     userbillingHistory,
				csrf.TemplateTag:         csrf.TemplateField(r),
			}

			t := template.New("billing.html")
			t, _ = t.ParseFiles("billing/billing.html")
			t.Execute(w, data)
		} else {
			// If the user does not have billing profile that means that they
			// have not started their trial yet.
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}
	}
}

func PaymentMethodsPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		user, err := apiControllers.GetCurrentUser(c, r)

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			http.Redirect(w, r, "https://tabulae.newsai.co/", 302)
			return
		}

		userBilling, err := apiControllers.GetUserBilling(c, r, user)

		// If the user has a billing profile
		if err == nil {
			cards, err := billing.GetUserCards(r, user, &userBilling)
			if err != nil {
				cards = []billing.Card{}
			}

			userFullName := strings.Join([]string{user.FirstName, user.LastName}, " ")

			data := map[string]interface{}{
				"userEmail":      user.Email,
				"userCards":      cards,
				"userFullName":   userFullName,
				"cardsOnFile":    len(userBilling.CardsOnFile),
				csrf.TemplateTag: csrf.TemplateField(r),
			}

			t := template.New("payments.html")
			t, _ = t.ParseFiles("billing/payments.html")
			t.Execute(w, data)

		} else {
			// If the user does not have billing profile that means that they
			// have not started their trial yet.
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}
	}
}

func PaymentMethodsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		user, err := apiControllers.GetCurrentUser(c, r)

		stripeToken := r.FormValue("stripeToken")

		if r.URL.Query().Get("next") != "" {
			session, _ := Store.Get(r, "sess")
			session.Values["next"] = r.URL.Query().Get("next")
			session.Save(r, w)

			// If there is a next and the user has not been logged in
			if err != nil {
				http.Redirect(w, r, r.URL.Query().Get("next"), 302)
				return
			}
		}

		// If there is no next and the user is not logged in
		if err != nil {
			http.Redirect(w, r, "https://tabulae.newsai.co/", 302)
			return
		}

		userBilling, err := apiControllers.GetUserBilling(c, r, user)
		if err != nil {
			log.Errorf(c, "%v", err)
			http.Redirect(w, r, "/api/billing/plans/trial", 302)
			return
		}

		err = billing.AddPaymentsToCustomer(r, user, &userBilling, stripeToken)

		// Throw error message to user
		if err != nil {
			log.Errorf(c, "%v", err)
			http.Redirect(w, r, "/api/billing/payment-methods?error="+err.Error(), 302)
			return
		}

		http.Redirect(w, r, "/api/billing/payment-methods", 302)
		return
	}
}
