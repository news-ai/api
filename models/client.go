package models

import (
	"net/http"
	"time"

	"google.golang.org/appengine/log"

	"golang.org/x/net/context"

	"github.com/qedus/nds"
)

type Client struct {
	Base

	Name  string   `json:"name"`
	URL   string   `json:"url"`
	Notes string   `json:"notes"`
	Tags  []string `json:"tags"`

	TeamId int64 `json:"teamid"`

	LinkedIn  string   `json:"linkedin"`
	Twitter   string   `json:"twitter"`
	Instagram string   `json:"instagram"`
	Websites  []string `json:"websites"`
	Blog      string   `json:"blog"`
}

/*
* Public methods
 */

/*
* Create methods
 */

func (cl *Client) Create(c context.Context, r *http.Request, currentUser User) (*Client, error) {
	cl.CreatedBy = currentUser.Id
	cl.Created = time.Now()
	_, err := cl.Save(c)
	return cl, err
}

/*
* Update methods
 */

// Function to save a new billing into App Engine
func (cl *Client) Save(c context.Context) (*Client, error) {
	// Update the Updated time
	cl.Updated = time.Now()

	k, err := nds.Put(c, cl.BaseKey(c, "Client"), cl)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}
	cl.Id = k.IntID()
	return cl, nil
}
