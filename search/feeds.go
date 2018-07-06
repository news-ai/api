package search

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"

	apiModels "github.com/news-ai/api/models"

	elastic "github.com/news-ai/elastic-appengine"
	"github.com/news-ai/tabulae/models"
)

var (
	elasticFeed *elastic.Elastic
)

type Feed struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdat"`

	Title         string `json:"title"`
	Author        string `json:"author"`
	Url           string `json:"url"`
	Summary       string `json:"summary"`
	FeedURL       string `json:"feedurl"`
	PublicationId int64  `json:"publicationid"`

	// Shared between tweet and instagram post
	Text string `json:"text"`

	TweetId         int64  `json:"tweetid"`
	TweetIdStr      string `json:"tweetidstr"`
	Username        string `json:"username"`
	TwitterLikes    int    `json:"twitterlikes"`
	TwitterRetweets int    `json:"twitterretweets"`

	InstagramUsername string `json:"instagramusername"`
	InstagramId       string `json:"instagramid"`
	InstagramImage    string `json:"instagramimage"`
	InstagramVideo    string `json:"instagramvideo"`
	InstagramLink     string `json:"instagramlink"`
	InstagramLikes    int    `json:"instagramlikes"`
	InstagramComments int    `json:"instagramcomments"`
	InstagramWidth    int    `json:"instagramwidth"`
	InstagramHeight   int    `json:"instagramheight"`
}

func (f *Feed) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := apiModels.SetField(f, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchFeed(c context.Context, elasticQuery interface{}, contacts []models.Contact, feedUrls []models.Feed) ([]Feed, int, error) {
	hits, err := elasticFeed.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Feed{}, 0, err
	}

	feedsMap := map[string]bool{}
	for i := 0; i < len(feedUrls); i++ {
		feedsMap[strings.ToLower(feedUrls[i].FeedURL)] = true
	}

	twitterUsernamesMap := map[string]bool{}
	for i := 0; i < len(contacts); i++ {
		twitterUsernamesMap[strings.ToLower(contacts[i].Twitter)] = true
	}

	instagramUsernamesMap := map[string]bool{}
	for i := 0; i < len(contacts); i++ {
		instagramUsernamesMap[strings.ToLower(contacts[i].Instagram)] = true
	}

	feedHits := hits.Hits
	feeds := []Feed{}
	for i := 0; i < len(feedHits); i++ {
		rawFeed := feedHits[i].Source.Data
		rawMap := rawFeed.(map[string]interface{})
		feed := Feed{}
		err := feed.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		feed.Type = strings.ToLower(feed.Type)
		feed.Type += "s"

		if feed.FeedURL != "" {
			// Reverse the #40 that we encoded it with
			if _, ok := feedsMap[strings.ToLower(feed.FeedURL)]; !ok {
				continue
			}
		} else {
			if feed.Type == "tweets" {
				if _, ok := twitterUsernamesMap[strings.ToLower(feed.Username)]; !ok {
					continue
				}
			} else if feed.Type == "instagrams" {
				if _, ok := instagramUsernamesMap[strings.ToLower(feed.InstagramUsername)]; !ok {
					continue
				}
			}
		}

		feeds = append(feeds, feed)
	}

	return feeds, hits.Total, nil
}

func SearchFeedForContacts(c context.Context, r *http.Request, contacts []models.Contact, feeds []models.Feed) ([]Feed, int, error) {
	// If contacts or feeds are empty return right away
	if len(contacts) == 0 && len(feeds) == 0 {
		return []Feed{}, 0, nil
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticFilterWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	for i := 0; i < len(contacts); i++ {
		if contacts[i].Twitter != "" {
			elasticUsernameQuery := ElasticUsernameQuery{}
			elasticUsernameQuery.Term.Username = strings.ToLower(contacts[i].Twitter)
			elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticUsernameQuery)
		}

		if contacts[i].Instagram != "" {
			elasticInstagramUsernameQuery := ElasticInstagramUsernameQuery{}
			elasticInstagramUsernameQuery.Term.InstagramUsername = strings.ToLower(contacts[i].Instagram)
			elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticInstagramUsernameQuery)
		}
	}

	for i := 0; i < len(feeds); i++ {
		if feeds[i].FeedURL != "" {
			elasticFeedUrlQuery := ElasticFeedUrlQuery{}
			elasticFeedUrlQuery.Match.FeedURL = strings.ToLower(feeds[i].FeedURL)
			elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticFeedUrlQuery)
		}
	}

	if len(elasticQuery.Query.Bool.Should) == 0 {
		return []Feed{}, 0, nil
	}

	minMatch := "50%"
	if len(elasticQuery.Query.Bool.Should) > 2 {
		approxMatch := float64(100 / len(elasticQuery.Query.Bool.Should))
		minMatch = fmt.Sprint(approxMatch) + "%"
	}

	minScore := float32(0.2)
	if len(elasticQuery.Query.Bool.Should) == 1 {
		minScore = float32(1.0)
	}

	if len(elasticQuery.Query.Bool.Should) > 10 {
		minScore = float32(0.1)
	}

	if len(elasticQuery.Query.Bool.Should) > 20 {
		minScore = float32(0.0)
	}

	elasticQuery.Query.Bool.MinimumShouldMatch = minMatch
	elasticQuery.MinScore = minScore

	elasticCreatedAtQuery := ElasticSortDataCreatedAtQuery{}
	elasticCreatedAtQuery.DataCreatedAt.Order = "desc"
	elasticCreatedAtQuery.DataCreatedAt.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedAtQuery)

	return searchFeed(c, elasticQuery, contacts, feeds)
}
