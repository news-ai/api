package controllers

import (
	"errors"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	gcontext "github.com/gorilla/context"
	"github.com/qedus/nds"

	"github.com/news-ai/web/utilities"

	"github.com/news-ai/api/models"
	"github.com/news-ai/tabulae/search"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getAgency(c context.Context, id int64) (models.Agency, error) {
	if id == 0 {
		return models.Agency{}, errors.New("datastore: no such entity")
	}

	// Get the agency by id
	var agency models.Agency
	agencyId := datastore.NewKey(c, "Agency", "", id, nil)
	err := nds.Get(c, agencyId, &agency)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Agency{}, err
	}

	if !agency.Created.IsZero() {
		agency.Format(agencyId, "agencies")
		return agency, nil
	}

	return models.Agency{}, errors.New("No agency by this id")
}

/*
* Filter methods
 */

func filterAgency(c context.Context, queryType, query string) (models.Agency, error) {
	ks, err := datastore.NewQuery("Agency").Filter(queryType+" =", query).KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Agency{}, err
	}

	if len(ks) == 0 {
		return models.Agency{}, errors.New("No agency by the field " + queryType)
	}

	var agencies []models.Agency
	agencies = make([]models.Agency, len(ks))
	err = nds.GetMulti(c, ks, agencies)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Agency{}, err
	}

	if len(agencies) > 0 {
		agencies[0].Format(ks[0], "agencies")
		return agencies[0], nil
	}

	return models.Agency{}, errors.New("No agency by the field " + queryType)
}

/*
* Public methods
 */

/*
* Get methods
 */

// Gets every single agency
func GetAgencies(c context.Context, r *http.Request) ([]models.Agency, interface{}, int, int, error) {
	// If user is querying then it is not denied by the server
	queryField := gcontext.Get(r, "q").(string)
	if queryField != "" {
		agencies, total, err := search.SearchAgency(c, r, queryField)
		if err != nil {
			return []models.Agency{}, nil, 0, 0, err
		}
		return agencies, nil, len(agencies), total, nil
	}

	// Now if user is not querying then check
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Agency{}, nil, 0, 0, err
	}

	if !user.IsAdmin {
		return []models.Agency{}, nil, 0, 0, errors.New("Forbidden")
	}

	query := datastore.NewQuery("Agency")
	query = ConstructQuery(query, r)

	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Agency{}, nil, 0, 0, err
	}

	var agencies []models.Agency
	agencies = make([]models.Agency, len(ks))
	err = nds.GetMulti(c, ks, agencies)
	if err != nil {
		log.Infof(c, "%v", err)
		return agencies, nil, 0, 0, err
	}

	for i := 0; i < len(agencies); i++ {
		agencies[i].Format(ks[i], "agencies")
	}

	return agencies, nil, len(agencies), 0, nil
}

func GetAgency(c context.Context, id string) (models.Agency, interface{}, error) {
	// Get the details of the current agency
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Agency{}, nil, err
	}

	agency, err := getAgency(c, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Agency{}, nil, err
	}

	return agency, nil, nil
}

/*
* Create methods
 */

func CreateAgencyFromUser(c context.Context, r *http.Request, u *models.User) (models.Agency, error) {
	agencyEmail, err := utilities.ExtractEmailExtension(u.Email)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Agency{}, err
	}

	agency, err := FilterAgencyByEmail(c, agencyEmail)
	if err != nil {
		agency = models.Agency{}
		agency.Name, err = utilities.ExtractNameFromEmail(agencyEmail)
		agency.Email = agencyEmail
		agency.Created = time.Now()

		// The person who signs up for the agency at the beginning
		// becomes the defacto administrator until we change.
		agency.Administrators = append(agency.Administrators, u.Id)
		currentUser, err := GetCurrentUser(c, r)
		if err != nil {
			log.Errorf(c, "%v", err)
			return agency, err
		}

		agency.Create(c, r, currentUser)
	}

	u.Employers = append(u.Employers, agency.Id)
	u.Save(c)
	return agency, nil
}

/*
* Filter methods
 */

func FilterAgencyByEmail(c context.Context, email string) (models.Agency, error) {
	// Get the id of the current agency
	agency, err := filterAgency(c, "Email", email)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Agency{}, err
	}

	return agency, nil
}
