package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	gcontext "github.com/gorilla/context"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/api/billing"

	"github.com/news-ai/api/models"
	"github.com/news-ai/tabulae/emails"
	"github.com/news-ai/tabulae/sync"

	"github.com/news-ai/web/permissions"
	"github.com/news-ai/web/utilities"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getUser(c context.Context, r *http.Request, id int64) (models.User, error) {
	// Get the current signed in user details by Id
	var user models.User
	userId := datastore.NewKey(c, "User", "", id, nil)
	err := nds.Get(c, userId, &user)

	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}

	if user.Email != "" {
		user.Format(userId, "users")
		currentUser, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, err
		}

		if user.TeamId != currentUser.TeamId && !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
			err = errors.New("Forbidden")
			log.Errorf(c, "%v", err)
			return models.User{}, err
		}

		return user, nil
	}
	return models.User{}, errors.New("No user by this id")
}

func getUserUnauthorized(c context.Context, r *http.Request, id int64) (models.User, error) {
	// Get the current signed in user details by Id
	var user models.User
	userId := datastore.NewKey(c, "User", "", id, nil)
	err := nds.Get(c, userId, &user)

	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}

	if user.Email != "" {
		user.Format(userId, "users")
		return user, nil
	}
	return models.User{}, errors.New("No user by this id")
}

// Gets every single user
func getUsers(c context.Context, r *http.Request) ([]models.User, error) {
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.User{}, err
	}

	if !user.IsAdmin {
		return []models.User{}, errors.New("Forbidden")
	}

	query := datastore.NewQuery("User")
	query = ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.User{}, err
	}

	var users []models.User
	users = make([]models.User, len(ks))
	err = nds.GetMulti(c, ks, users)
	if err != nil {
		log.Infof(c, "%v", err)
		return users, err
	}

	for i := 0; i < len(users); i++ {
		users[i].Format(ks[i], "users")
	}
	return users, nil
}

// Gets every single user
func getUsersUnauthorized(c context.Context, r *http.Request) ([]models.User, error) {
	query := datastore.NewQuery("User")
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.User{}, err
	}

	var users []models.User
	users = make([]models.User, len(ks))
	err = nds.GetMulti(c, ks, users)
	if err != nil {
		log.Infof(c, "%v", err)
		return users, err
	}

	for i := 0; i < len(users); i++ {
		users[i].Format(ks[i], "users")
	}
	return users, nil
}

/*
* Filter methods
 */

func filterUser(c context.Context, queryType, query string) (models.User, error) {
	// Get the current signed in user details by Id
	ks, err := datastore.NewQuery("User").Filter(queryType+" =", query).Limit(1).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}

	if len(ks) == 0 {
		return models.User{}, errors.New("No user by the field " + queryType)
	}

	user := models.User{}
	userId := ks[0]

	err = nds.Get(c, userId, &user)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}

	if !user.Created.IsZero() {
		user.Format(userId, "users")
		return user, nil
	}
	return models.User{}, errors.New("No user by this " + queryType)
}

func filterUserConfirmed(c context.Context, queryType, query string) (models.User, error) {
	// Get the current signed in user details by Id
	ks, err := datastore.NewQuery("User").Filter(queryType+" =", query).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}

	if len(ks) == 0 {
		return models.User{}, errors.New("No user by the field " + queryType)
	}

	// This shouldn't happen, but if the user double registers. Handle this case I guess
	if len(ks) > 1 {
		whichUserConfirmed := models.User{}
		for i := 0; i < len(ks); i++ {
			user := models.User{}
			userId := ks[i]

			err = nds.Get(c, userId, &user)
			if err != nil {
				log.Errorf(c, "%v", err)
				return models.User{}, err
			}

			if !user.Created.IsZero() {
				user.Format(userId, "users")

				if user.EmailConfirmed {
					whichUserConfirmed = user
				}
			}
		}

		// If none of them have confirmed their emails
		if whichUserConfirmed.Email == "" {
			user := models.User{}
			userId := ks[0]

			err = nds.Get(c, userId, &user)
			if err != nil {
				log.Errorf(c, "%v", err)
				return models.User{}, err
			}

			if !user.Created.IsZero() {
				user.Format(userId, "users")
				return user, nil
			}
		}

		return whichUserConfirmed, nil
	} else {
		// The normal case where there's only one email
		user := models.User{}
		userId := ks[0]

		err = nds.Get(c, userId, &user)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, err
		}

		if !user.Created.IsZero() {
			user.Format(userId, "users")
			return user, nil
		}
	}

	return models.User{}, errors.New("No user by this " + queryType)
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetUsers(c context.Context, r *http.Request) ([]models.User, interface{}, int, int, error) {
	// Get the current user
	users, err := getUsers(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.User{}, nil, 0, 0, err
	}

	return users, nil, len(users), 0, nil
}

func GetUsersUnauthorized(c context.Context, r *http.Request) ([]models.User, error) {
	// Get the current user
	users, err := getUsersUnauthorized(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.User{}, err
	}

	return users, nil
}

func GetUserById(c context.Context, r *http.Request, id int64) (models.User, interface{}, error) {
	// Get the details of the current user
	user, err := getUser(c, r, id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}
	return user, nil, nil
}

func GetUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	// Get the details of the current user
	switch id {
	case "me":
		user, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		return user, nil, err
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err := getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		return user, nil, nil
	}
}

func GetUserByEmailForValidation(c context.Context, email string) (models.User, error) {
	// Get the current user
	user, err := filterUserConfirmed(c, "Email", email)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}
	return user, nil
}

func GetUserByEmail(c context.Context, email string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "Email", email)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}
	return user, nil
}

func GetUserByApiKey(c context.Context, apiKey string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "ApiKey", apiKey)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}
	return user, nil
}

func GetUserByConfirmationCode(c context.Context, confirmationCode string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "ConfirmationCode", confirmationCode)
	if err != nil {
		log.Errorf(c, "%v", err)

		// Have a backup ConfirmationCode
		userBackup, errBackup := filterUser(c, "ConfirmationCodeBackup", confirmationCode)
		if errBackup != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, err
		}

		return userBackup, nil
	}
	return user, nil
}

func GetUserByResetCode(c context.Context, resetCode string) (models.User, error) {
	// Get the current user
	user, err := filterUser(c, "ResetPasswordCode", resetCode)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}
	return user, nil
}

func GetCurrentUser(c context.Context, r *http.Request) (models.User, error) {
	// Get the current user
	_, ok := gcontext.GetOk(r, "user")
	if !ok {
		return models.User{}, errors.New("No user logged in")
	}
	user := gcontext.Get(r, "user").(models.User)
	return user, nil
}

func GetUserFromApiKey(r *http.Request, ApiKey string) (models.User, error) {
	c := appengine.NewContext(r)
	user, err := GetUserByApiKey(c, ApiKey)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, err
	}
	return user, nil
}

func GetUserByIdUnauthorized(c context.Context, r *http.Request, userId int64) (models.User, error) {
	// Method dangerous since it can log into as any user. Be careful.
	user, err := getUserUnauthorized(c, r, userId)
	if err != nil {
		return models.User{}, nil
	}
	return user, nil
}

func AddUserToContext(c context.Context, r *http.Request, email string) {
	_, ok := gcontext.GetOk(r, "user")
	if !ok {
		user, _ := GetUserByEmail(c, email)
		gcontext.Set(r, "user", user)
		Update(c, r, &user)
	} else {
		user := gcontext.Get(r, "user").(models.User)
		Update(c, r, &user)
	}
}

func AddPlanToUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	if !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var userNewPlan models.UserNewPlan
	err = decoder.Decode(buf, &userNewPlan)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	userBilling, err := GetUserBilling(c, r, user)

	if len(userBilling.CardsOnFile) == 0 {
		return user, userBilling, errors.New("This user has no cards on file")
	}

	originalPlan := ""
	switch userNewPlan.Plan {
	case "bronze":
		originalPlan = "Personal"
	case "silver-1":
		originalPlan = "Freelancer"
	case "gold-1":
		originalPlan = "Business"
	}

	if userNewPlan.Duration != "monthly" && userNewPlan.Duration != "annually" {
		return user, userBilling, errors.New("Duration is invalid")
	}

	if userNewPlan.Plan == "" {
		return user, userBilling, errors.New("Plan is invalid")
	}

	if originalPlan == "" {
		return user, userBilling, errors.New("Original Plan is invalid")
	}

	err = billing.AddPlanToUser(r, user, &userBilling, userNewPlan.Plan, userNewPlan.Duration, userNewPlan.Coupon, originalPlan)
	if err != nil {
		log.Errorf(c, "%v", err)
		return user, userBilling, err
	}

	return user, userBilling, nil
}

func AddEmailToUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	// Only available when using SendGrid
	if user.Gmail || user.ExternalEmail {
		return user, nil, errors.New("Feature only works when using Sendgrid")
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var userEmail models.UserEmail
	err = decoder.Decode(buf, &userEmail)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	userEmail.Email = strings.ToLower(userEmail.Email)
	validEmail, err := mail.ParseAddress(userEmail.Email)
	if err != nil {
		return user, nil, err
	}

	if user.Email == validEmail.Address {
		return user, nil, errors.New("Can't add your default email as an extra email")
	}

	for i := 0; i < len(user.Emails); i++ {
		if user.Emails[i] == validEmail.Address {
			return user, nil, errors.New("Email already exists for the user")
		}
	}

	// Generate User Emails Code to send to confirmation email
	userEmailCode := models.UserEmailCode{}
	userEmailCode.InviteCode = utilities.RandToken()
	userEmailCode.Email = validEmail.Address
	userEmailCode.Create(c, r, currentUser)

	// Send Confirmation Email to this email address
	addUserEmailErr := emails.AddEmailToUser(c, user, validEmail.Address, userEmailCode.InviteCode)
	if addUserEmailErr != nil {
		return user, nil, err
	}

	return user, nil, nil
}

func RemoveEmailFromUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var userEmail models.UserEmail
	err = decoder.Decode(buf, &userEmail)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	userEmail.Email = strings.ToLower(userEmail.Email)
	validEmail, err := mail.ParseAddress(userEmail.Email)
	if err != nil {
		return user, nil, err
	}

	if user.Email == validEmail.Address {
		return user, nil, errors.New("Can't remove your default email as an extra email")
	}

	for i := 0; i < len(user.Emails); i++ {
		if user.Emails[i] == validEmail.Address {
			user.Emails = append(user.Emails[:i], user.Emails[i+1:]...)
		}
	}

	SaveUser(c, r, &user)
	return user, nil, nil
}

func GetUserDailyEmail(c context.Context, r *http.Request, user models.User) int {
	t := time.Now()
	todayDateMorning := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	todayDateNight := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 59, time.Local)

	emailsSent, err := datastore.NewQuery("Email").Filter("CreatedBy =", user.Id).Filter("IsSent =", true).Filter("Cancel =", false).Filter("Created <=", todayDateNight).Filter("Created >=", todayDateMorning).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return 0
	}

	return len(emailsSent)
}

func GetUserPlanDetails(c context.Context, r *http.Request, id string) (models.UserPlan, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.UserPlan{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.UserPlan{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.UserPlan{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserPlan{}, nil, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return models.UserPlan{}, nil, err
	}

	userBilling, err := GetUserBilling(c, r, user)
	if err != nil {
		return models.UserPlan{}, nil, err
	}

	userPlan := models.UserPlan{}
	userPlanName := billing.BillingIdToPlanName(userBilling.StripePlanId)
	userPlan.PlanName = userPlanName
	userPlan.EmailAccounts = billing.UserMaximumEmailAccounts(userPlanName)
	userPlan.OnTrial = userBilling.IsOnTrial
	userPlan.DailyEmailsAllowed = billing.StripePlanIdToMaximumEmailSent(userBilling.StripePlanId)
	userPlan.EmailsSentToday = GetUserDailyEmail(c, r, user)

	return userPlan, nil, nil
}

func ConfirmAddEmailToUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return user, nil, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return user, nil, err
	}

	if r.URL.Query().Get("code") != "" {
		query := datastore.NewQuery("UserEmailCode").Filter("InviteCode =", r.URL.Query().Get("code"))
		query = ConstructQuery(query, r)
		ks, err := query.KeysOnly().GetAll(c, nil)
		if err != nil {
			log.Errorf(c, "%v", err)
			return user, nil, err
		}

		var userEmailCodes []models.UserEmailCode
		userEmailCodes = make([]models.UserEmailCode, len(ks))
		err = nds.GetMulti(c, ks, userEmailCodes)
		if err != nil {
			log.Errorf(c, "%v", err)
			return user, nil, err
		}

		if len(userEmailCodes) > 0 {
			if !permissions.AccessToObject(user.Id, userEmailCodes[0].CreatedBy) {
				err = errors.New("Forbidden")
				log.Errorf(c, "%v", err)
				return user, nil, err
			}
			alreadyExists := false
			for i := 0; i < len(user.Emails); i++ {
				if user.Emails[i] == userEmailCodes[0].Email {
					alreadyExists = true
				}
			}

			if !alreadyExists {
				user.Emails = append(user.Emails, userEmailCodes[0].Email)
				SaveUser(c, r, &user)
			}
			return user, nil, nil
		}

		return user, nil, errors.New("No code by the code you entered")
	}

	return user, nil, errors.New("No code present")
}

func FeedbackFromUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var userFeedback models.UserFeedback
	err = decoder.Decode(buf, &userFeedback)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	// Get user's billing profile and add reasons there
	userBilling, err := GetUserBilling(c, r, currentUser)
	userBilling.ReasonNotPurchase = userFeedback.ReasonNotPurchase
	userBilling.FeedbackAfterTrial = userFeedback.FeedbackAfterTrial
	userBilling.Save(c)

	// Set the trial feedback to true - since they gave us feedback now
	user.TrialFeedback = true
	user.Save(c)

	sync.ResourceSync(r, user.Id, "User", "create")
	return user, nil, nil
}

/*
* Update methods
 */

func SaveUser(c context.Context, r *http.Request, u *models.User) (*models.User, error) {
	u.Save(c)
	sync.ResourceSync(r, u.Id, "User", "create")
	return u, nil
}

func Update(c context.Context, r *http.Request, u *models.User) (*models.User, error) {
	if len(u.Employers) == 0 {
		CreateAgencyFromUser(c, r, u)
	}

	billing, err := GetUserBilling(c, r, *u)
	if err != nil {
		return u, err
	}

	if billing.Expires.Before(time.Now()) {
		if billing.IsOnTrial {
			u.IsActive = false
			u.Save(c)

			billing.IsOnTrial = false
			billing.Save(c)
		} else {
			if billing.IsCancel {
				u.IsActive = false
				u.Save(c)
			} else {
				if billing.StripePlanId != "free" {
					// If they haven't canceled then we can add a month until they do.
					// More sophisticated to add the amount depending on what
					// plan they were on.
					addAMonth := billing.Expires.AddDate(0, 1, 0)
					billing.Expires = addAMonth
					billing.Save(c)

					// Keep the user active
					u.IsActive = true
					u.Save(c)
				}
			}
		}
	}

	return u, nil
}

func UpdateUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var updatedUser models.User
	err = decoder.Decode(buf, &updatedUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	utilities.UpdateIfNotBlank(&user.FirstName, updatedUser.FirstName)
	utilities.UpdateIfNotBlank(&user.LastName, updatedUser.LastName)
	utilities.UpdateIfNotBlank(&user.EmailSignature, updatedUser.EmailSignature)

	// If new user wants to get daily emails
	if updatedUser.GetDailyEmails == true {
		user.GetDailyEmails = true
	}

	// If this person doesn't want to get daily emails anymore
	if user.GetDailyEmails == true && updatedUser.GetDailyEmails == false {
		user.GetDailyEmails = false
	}

	if user.SMTPValid {
		// If new user wants to get daily emails
		if updatedUser.ExternalEmail == true {
			user.ExternalEmail = true
		}

		// If this person doesn't want to get daily emails anymore
		if user.ExternalEmail == true && updatedUser.ExternalEmail == false {
			user.ExternalEmail = false
		}
	}

	if len(updatedUser.Employers) > 0 {
		user.Employers = updatedUser.Employers
	}

	if len(updatedUser.EmailSignatures) > 0 {
		user.EmailSignatures = updatedUser.EmailSignatures
	}

	// Special case when you want to remove all the email signatures
	if len(user.EmailSignatures) > 0 && len(updatedUser.EmailSignatures) == 0 {
		user.EmailSignatures = updatedUser.EmailSignatures
	}

	user.Save(c)
	sync.ResourceSync(r, user.Id, "User", "create")
	return user, nil, nil
}

/*
* Action methods
 */

func BanUser(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	user.IsActive = false
	user.IsBanned = true
	SaveUser(c, r, &user)
	return user, nil, nil
}

func GetAndRefreshLiveToken(c context.Context, r *http.Request, id string) (interface{}, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	token := models.UserLiveToken{}
	token.Token = user.LiveAccessToken
	token.Expires = user.LiveAccessTokenExpire
	return token, nil, nil
}

func ValidateUserPassword(r *http.Request, email string, password string) (models.User, bool, error) {
	c := appengine.NewContext(r)
	user, err := GetUserByEmailForValidation(c, email)
	if err == nil {
		err = utilities.ValidatePassword(user.Password, password)
		if err != nil {
			log.Errorf(c, "%v", err)
			return user, false, nil
		}
		return user, true, nil
	}
	return models.User{}, false, errors.New("User does not exist")
}

func SetUser(c context.Context, r *http.Request, userId int64) (models.User, error) {
	// Method dangerous since it can log into as any user. Be careful.
	user, err := getUserUnauthorized(c, r, userId)
	if err != nil {
		log.Errorf(c, "%v", err)
	}
	gcontext.Set(r, "user", user)
	return user, nil
}

func UpdateUserEmail(c context.Context, r *http.Request, id string) (models.User, interface{}, error) {
	user := models.User{}
	err := errors.New("")

	switch id {
	case "me":
		user, err = GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	default:
		userId, err := utilities.StringIdToInt(id)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
		user, err = getUser(c, r, userId)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.User{}, nil, err
		}
	}

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	if !permissions.AccessToObject(user.Id, currentUser.Id) && !currentUser.IsAdmin {
		err = errors.New("Forbidden")
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var updatedUser models.User
	err = decoder.Decode(buf, &updatedUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.User{}, nil, err
	}

	// If new user wants to get daily emails
	if updatedUser.Email != "" {
		user.Email = updatedUser.Email
	}

	user.Save(c)
	sync.ResourceSync(r, user.Id, "User", "create")
	return user, nil, nil
}
