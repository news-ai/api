package controllers

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

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

func getClient(c context.Context, id int64) (models.Client, error) {
	if id == 0 {
		return models.Client{}, errors.New("datastore: no such entity")
	}

	// Get the client details by id
	var client models.Client
	clientId := datastore.NewKey(c, "Client", "", id, nil)

	err := nds.Get(c, clientId, &client)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Client{}, err
	}

	if !client.Created.IsZero() {
		client.Format(clientId, "clients")
		return client, nil
	}
	return models.Client{}, errors.New("No client by this id")
}

/*
* Public methods
 */

/*
* Get methods
 */

func GetClients(c context.Context, r *http.Request) ([]models.Client, interface{}, int, int, error) {
	// Now if user is not querying then check
	user, err := GetCurrentUser(c, r)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Client{}, nil, 0, 0, err
	}

	if !user.IsAdmin {
		return []models.Client{}, nil, 0, 0, errors.New("Forbidden")
	}

	query := datastore.NewQuery("Client")
	query = ConstructQuery(query, r)
	ks, err := query.KeysOnly().GetAll(c, nil)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Client{}, nil, 0, 0, err
	}

	var clients []models.Client
	clients = make([]models.Client, len(ks))
	err = nds.GetMulti(c, ks, clients)
	if err != nil {
		log.Infof(c, "%v", err)
		return clients, nil, 0, 0, err
	}

	for i := 0; i < len(clients); i++ {
		clients[i].Format(ks[i], "clients")
	}

	return clients, nil, len(clients), 0, nil
}

func GetClient(c context.Context, id string) (models.Client, interface{}, error) {
	// Get the details of a client
	currentId, err := utilities.StringIdToInt(id)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Client{}, nil, err
	}

	client, err := getClient(c, currentId)
	if err != nil {
		log.Errorf(c, "%v", err)
		return models.Client{}, nil, err
	}
	return client, nil, nil
}
