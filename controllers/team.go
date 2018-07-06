package controllers

import (
	"errors"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/qedus/nds"

	"github.com/news-ai/api/models"

	"github.com/news-ai/web/utilities"
)

/*
* Private methods
 */

/*
* Get methods
 */

func getTeam(c context.Context, id int64) (models.Team, error) {
	if id == 0 {
		return models.Team{}, errors.New("datastore: no such entity")
	}

	// Get the team details by id
	var team models.Team
	teamId := datastore.NewKey(c, "Team", "", id, nil)

	err := nds.Get(c, teamId, &team)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Team{}, err
	}

	if !team.Created.IsZero() {
		team.Format(teamId, "teams")
		return team, nil
	}
	return models.Team{}, errors.New("No team by this id")
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetTeams(c context.Context, r *http.Request) ([]models.Team, interface{}, int, int, error) {
	// Now if user is not querying then check
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Team{}, nil, 0, 0, err
	}

	if !user.IsAdmin {
		return []models.Team{}, nil, 0, 0, errors.New("Forbidden")
	}

	query := datastore.NewQuery("Team")
	query = ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Team{}, nil, 0, 0, err
	}

	var teams []models.Team
	teams = make([]models.Team, len(ks))
	err = nds.GetMulti(c, ks, teams)
	if err != nil {
		log.Infof(c, "%v", err)
		return teams, nil, 0, 0, err
	}

	for i := 0; i < len(teams); i++ {
		teams[i].Format(ks[i], "teams")
	}

	return teams, nil, len(teams), 0, nil
}

func GetTeam(c context.Context, id string) (models.Team, interface{}, error) {
	// Get the details of the current team
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Team{}, nil, err
	}

	team, err := getTeam(c, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Team{}, nil, err
	}
	return team, nil, nil
}

/*
* Create methods
 */

func CreateTeam(c context.Context, r *http.Request) ([]models.Team, interface{}, error) {
	buf, _ := ioutil.ReadAll(r.Body)

	currentUser, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Team{}, nil, err
	}

	if !currentUser.IsAdmin {
		return []models.Team{}, nil, errors.New("Forbidden")
	}

	decoder := ffjson.NewDecoder()
	var team models.Team
	err = decoder.Decode(buf, &team)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Team{}, nil, err
	}

	if len(team.Members) > team.MaxMembers {
		return []models.Team{}, nil, errors.New("The number of members is greater than the allowed number of members")
	}

	// Create team
	_, err = team.Create(c, r, currentUser)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Team{}, nil, err
	}

	// Add team Id to team members
	// Validate member accounts
	confirmMembers := []int64{}
	for i := 0; i < len(team.Members); i++ {
		user, err := getUser(c, r, team.Members[i])
		if err == nil && user.TeamId == 0 {
			confirmMembers = append(confirmMembers, user.Id)
			user.TeamId = team.Id
			user.Save(c)
		}
	}

	// Validate admin accounts
	confirmAdmins := []int64{}
	for i := 0; i < len(team.Admins); i++ {
		user, err := getUser(c, r, team.Admins[i])
		if err == nil {
			confirmAdmins = append(confirmAdmins, user.Id)
		}
	}

	team.Members = confirmMembers
	team.Admins = confirmAdmins
	team.Save(c)

	return []models.Team{team}, nil, nil
}
