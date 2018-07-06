package search

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/log"

	apiModels "github.com/news-ai/api/models"

	elastic "github.com/news-ai/elastic-appengine"
)

var (
	elasticInstagramTimeseries *elastic.Elastic
	elasticTwitterTimeseries   *elastic.Elastic
)

type TwitterTimeseries struct {
	Username  string    `json:"Username"`
	CreatedAt time.Time `json:"CreatedAt"`
	Followers int       `json:"Followers"`
	Following int       `json:"Following"`
	Likes     int       `json:"Likes"`
	Retweets  int       `json:"Retweets"`
	Posts     int       `json:"Posts"`
}

type InstagramTimeseries struct {
	Username  string    `json:"Username"`
	CreatedAt time.Time `json:"CreatedAt"`
	Followers int       `json:"Followers"`
	Following int       `json:"Following"`
	Likes     int       `json:"Likes"`
	Comments  int       `json:"Comments"`
	Posts     int       `json:"Posts"`
}

func (tt *TwitterTimeseries) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := apiModels.SetField(tt, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (it *InstagramTimeseries) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := apiModels.SetField(it, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchTwitterTimeseries(c context.Context, elasticQuery interface{}) (interface{}, int, error) {
	hits, err := elasticTwitterTimeseries.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, err
	}

	profileHits := hits.Hits

	if len(profileHits) == 0 {
		log.Infof(c, "%v", profileHits)
		return nil, 0, errors.New("No Twitter profile for this username")
	}

	var interfaceSlice = make([]interface{}, len(profileHits))

	for i := 0; i < len(profileHits); i++ {
		interfaceSlice[i] = profileHits[i].Source.Data
	}

	return interfaceSlice, hits.Total, nil
}

func searchInstagramTimeseries(c context.Context, elasticQuery interface{}) (interface{}, int, error) {
	hits, err := elasticInstagramTimeseries.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, err
	}

	profileHits := hits.Hits

	if len(profileHits) == 0 {
		log.Infof(c, "%v", profileHits)
		return nil, 0, errors.New("No Instagram profile for this username")
	}

	var interfaceSlice = make([]interface{}, len(profileHits))

	for i := 0; i < len(profileHits); i++ {
		interfaceSlice[i] = profileHits[i].Source.Data
	}

	return interfaceSlice, hits.Total, nil
}

func searchInstagramTimeseriesByUsernames(c context.Context, elasticQuery interface{}) ([]InstagramTimeseries, error) {
	hits, err := elasticInstagramTimeseries.QueryStructMGet(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	instagramTimeseriesData := []InstagramTimeseries{}
	for i := 0; i < len(hits); i++ {
		if hits[i].Found {
			rawInstagramTimeseries := hits[i].Source.Data
			rawMap := rawInstagramTimeseries.(map[string]interface{})
			instagramTimeseries := InstagramTimeseries{}
			err := instagramTimeseries.FillStruct(rawMap)
			if err != nil {
				log.Errorf(c, "%v", err)
			}

			instagramTimeseriesData = append(instagramTimeseriesData, instagramTimeseries)
		}
	}

	return instagramTimeseriesData, nil
}

func searchTwitterTimeseriesByUsernames(c context.Context, elasticQuery interface{}) ([]TwitterTimeseries, error) {
	hits, err := elasticTwitterTimeseries.QueryStructMGet(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	twitterTimeseriesData := []TwitterTimeseries{}
	for i := 0; i < len(hits); i++ {
		if hits[i].Found {
			rawTwitterTimeseries := hits[i].Source.Data
			rawMap := rawTwitterTimeseries.(map[string]interface{})
			twitterTimeseries := TwitterTimeseries{}
			err := twitterTimeseries.FillStruct(rawMap)
			if err != nil {
				log.Errorf(c, "%v", err)
			}

			twitterTimeseriesData = append(twitterTimeseriesData, twitterTimeseries)
		}
	}

	return twitterTimeseriesData, nil
}

func SearchInstagramTimeseriesByUsernames(c context.Context, r *http.Request, usernames []string) ([]InstagramTimeseries, error) {
	if len(usernames) == 0 {
		return nil, nil
	}

	elasticQuery := ElasticMGetQuery{}

	for i := 0; i < len(usernames); i++ {
		if usernames[i] != "" {
			dateToday := time.Now().Format("2006-01-02")
			elasticQuery.Ids = append(elasticQuery.Ids, usernames[i]+"-"+dateToday)
		}
	}

	return searchInstagramTimeseriesByUsernames(c, elasticQuery)
}

func SearchInstagramTimeseriesByUsernamesWithDays(c context.Context, r *http.Request, usernames []string, days int) ([]InstagramTimeseries, error) {
	if len(usernames) == 0 {
		return nil, nil
	}

	elasticQuery := ElasticMGetQuery{}
	timeNow := time.Now()

	for i := 0; i < len(usernames); i++ {
		if usernames[i] != "" {
			for x := 0; x < days; x++ {
				dateFormatted := timeNow.AddDate(0, 0, -1*x).Format("2006-01-02")
				elasticQuery.Ids = append(elasticQuery.Ids, usernames[i]+"-"+dateFormatted)
			}
		}
	}

	return searchInstagramTimeseriesByUsernames(c, elasticQuery)
}

func SearchTwitterTimeseriesByUsernames(c context.Context, r *http.Request, usernames []string) ([]TwitterTimeseries, error) {
	if len(usernames) == 0 {
		return nil, nil
	}

	elasticQuery := ElasticMGetQuery{}

	for i := 0; i < len(usernames); i++ {
		if usernames[i] != "" {
			dateToday := time.Now().Format("2006-01-02")
			elasticQuery.Ids = append(elasticQuery.Ids, usernames[i]+"-"+dateToday)
		}
	}

	return searchTwitterTimeseriesByUsernames(c, elasticQuery)
}

func SearchTwitterTimeseriesByUsernamesWithDays(c context.Context, r *http.Request, usernames []string, days int) ([]TwitterTimeseries, error) {
	if len(usernames) == 0 {
		return nil, nil
	}

	elasticQuery := ElasticMGetQuery{}
	timeNow := time.Now()

	for i := 0; i < len(usernames); i++ {
		if usernames[i] != "" {
			for x := 0; x < days; x++ {
				dateFormatted := timeNow.AddDate(0, 0, -1*x).Format("2006-01-02")
				elasticQuery.Ids = append(elasticQuery.Ids, usernames[i]+"-"+dateFormatted)
			}
		}
	}

	return searchTwitterTimeseriesByUsernames(c, elasticQuery)
}

func SearchInstagramTimeseriesByUsername(c context.Context, r *http.Request, username string) (interface{}, int, error) {
	if username == "" {
		return nil, 0, errors.New("Contact does not have a instagram username")
	}

	offset := 0
	limit := 31

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticUsernameQuery := ElasticUsernameQuery{}
	elasticUsernameQuery.Term.Username = strings.ToLower(username)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticUsernameQuery)

	elasticCreatedAtDateQuery := ElasticSortDataCreatedAtQuery{}
	elasticCreatedAtDateQuery.DataCreatedAt.Order = "desc"
	elasticCreatedAtDateQuery.DataCreatedAt.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedAtDateQuery)

	return searchInstagramTimeseries(c, elasticQuery)
}

func SearchTwitterTimeseriesByUsername(c context.Context, r *http.Request, username string) (interface{}, int, error) {
	if username == "" {
		return nil, 0, errors.New("Contact does not have a twitter username")
	}

	offset := 0
	limit := 31

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticUsernameQuery := ElasticUsernameQuery{}
	elasticUsernameQuery.Term.Username = strings.ToLower(username)
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticUsernameQuery)

	elasticCreatedAtDateQuery := ElasticSortDataCreatedAtQuery{}
	elasticCreatedAtDateQuery.DataCreatedAt.Order = "desc"
	elasticCreatedAtDateQuery.DataCreatedAt.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedAtDateQuery)

	return searchTwitterTimeseries(c, elasticQuery)
}
