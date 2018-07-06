package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/mail"

	"golang.org/x/net/context"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/news-ai/api/models"

	"github.com/news-ai/tabulae/emails"

	"github.com/news-ai/web/utilities"
)

/*
* Private
 */

/*
* Private methods
 */

/*
* Get methods
 */

func generateTokenAndEmail(c context.Context, r *http.Request, invite models.Invite) (models.UserInviteCode, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, err
	}

	validEmail, err := mail.ParseAddress(invite.Email)
	if err != nil {
		invalidEmailError := errors.New("Email user has entered is incorrect")
		log.Errorf(c, "%v", invalidEmailError)
		return models.UserInviteCode{}, invalidEmailError
	}

	// Get the Contact by id
	ks, err := datastore.NewQuery("UserInviteCode").Filter("Email =", validEmail.Address).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, err
	}

	if len(ks) > 0 {
		hasBeenUsed := false

		var invites []models.UserInviteCode
		invites = make([]models.UserInviteCode, len(ks))
		err = nds.GetMulti(c, ks, invites)
		if err != nil {
			log.Errorf(c, "%v", err)
			return models.UserInviteCode{}, err
		}

		for i := 0; i < len(invites); i++ {
			invites[i].Format(ks[i], "invites")

			if invites[i].IsUsed {
				hasBeenUsed = true
			}
		}

		if hasBeenUsed {
			invitedAlreadyError := errors.New("User has already been invited to the NewsAI platform")
			log.Errorf(c, "%v", invitedAlreadyError)
			return models.UserInviteCode{}, invitedAlreadyError
		}
	}

	// Check if the user is already a part of the platform
	_, err = GetUserByEmail(c, validEmail.Address)
	if err == nil {
		userExistsError := errors.New("User already exists on the NewsAI platform")
		log.Errorf(c, "%v", userExistsError)
		return models.UserInviteCode{}, userExistsError
	}

	referralCode := models.UserInviteCode{}
	referralCode.Email = validEmail.Address
	referralCode.InviteCode = utilities.RandToken()
	_, err = referralCode.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, err
	}

	// Email this person with the referral code
	inviteUserEmailErr := emails.InviteUser(c, currentUser, validEmail.Address, referralCode.InviteCode, invite.PersonalNote)
	if inviteUserEmailErr != nil {
		// Redirect user back to login page
		log.Errorf(c, "%v", "Invite email was not sent for "+validEmail.Address)
		log.Errorf(c, "%v", err)
		inviteEmailError := errors.New("Could not send invite email. We'll fix this soon!")
		return models.UserInviteCode{}, inviteEmailError
	}

	return referralCode, nil
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetInvites(c context.Context, r *http.Request) ([]models.UserInviteCode, interface{}, int, int, error) {
	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.UserInviteCode{}, nil, 0, 0, err
	}

	ks, err := datastore.NewQuery("UserInviteCode").Filter("CreatedBy =", currentUser.Id).Filter("IsUsed =", true).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.UserInviteCode{}, nil, 0, 0, err
	}

	var userInviteCodes []models.UserInviteCode
	userInviteCodes = make([]models.UserInviteCode, len(ks))
	err = nds.GetMulti(c, ks, userInviteCodes)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.UserInviteCode{}, nil, 0, 0, err
	}

	for i := 0; i < len(userInviteCodes); i++ {
		userInviteCodes[i].Format(ks[i], "invites")
	}

	return userInviteCodes, nil, len(userInviteCodes), 0, nil
}

func GetInviteFromInvitationCode(c context.Context, r *http.Request, invitationCode string) (models.UserInviteCode, error) {
	ks, err := datastore.NewQuery("UserInviteCode").Filter("InviteCode =", invitationCode).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, err
	}

	if len(ks) == 0 {
		return models.UserInviteCode{}, errors.New("Wrong invitation code")
	}

	var userInviteCodes []models.UserInviteCode
	userInviteCodes = make([]models.UserInviteCode, len(ks))
	err = nds.GetMulti(c, ks, userInviteCodes)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, err
	}

	if len(userInviteCodes) > 0 {
		userInviteCodes[0].Format(ks[0], "invites")
		return userInviteCodes[0], nil
	}

	return models.UserInviteCode{}, errors.New("No invitation by that code")
}

/*
* Create methods
 */

func CreateInvite(c context.Context, r *http.Request) (models.UserInviteCode, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	decoder := ffjson.NewDecoder()
	var invite models.Invite
	err := decoder.Decode(buf, &invite)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.UserInviteCode{}, nil, err
	}

	userInvite, err := generateTokenAndEmail(c, r, invite)
	if err != nil {
		return models.UserInviteCode{}, nil, err
	}

	userInvite.Type = "invites"

	return userInvite, nil, nil
}
