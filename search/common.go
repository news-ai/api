package search

import (
	elastic "github.com/news-ai/elastic-appengine"
)

var (
	NewBaseURL = "https://search.newsai.org"
)

type ElasticMGetQuery struct {
	Ids []string `json:"ids"`
}

type ElasticCreatedByQuery struct {
	Term struct {
		CreatedBy int64 `json:"data.CreatedBy"`
	} `json:"term"`
}

type ElasticSubjectQuery struct {
	Term struct {
		Subject string `json:"data.Subject"`
	} `json:"term"`
}

type ElasticBaseSubjectQuery struct {
	Term struct {
		BaseSubject string `json:"data.BaseSubject"`
	} `json:"term"`
}

type ElasticLocationCityQuery struct {
	Term struct {
		City string `json:"data.demographics.locationDeduced.city.name"`
	} `json:"term"`
}

type ElasticOrganizationNameQuery struct {
	Match struct {
		Name string `json:"data.organizations.name"`
	} `json:"match"`
}

type ElasticLocationStateQuery struct {
	Term struct {
		State string `json:"data.demographics.locationDeduced.state.name"`
	} `json:"term"`
}

type ElasticLocationCountryQuery struct {
	Term struct {
		Country string `json:"data.demographics.locationDeduced.country.name"`
	} `json:"term"`
}

type ElasticBaseOpenedQuery struct {
	Term struct {
		Opened int64 `json:"data.Opened"`
	} `json:"term"`
}

type ElasticBaseClickedQuery struct {
	Term struct {
		Clicked int64 `json:"data.Clicked"`
	} `json:"term"`
}

type ElasticIsSentQuery struct {
	Term struct {
		IsSent bool `json:"data.IsSent"`
	} `json:"term"`
}

type ElasticIsDeletedQuery struct {
	Term struct {
		IsDeleted bool `json:"data.IsDeleted"`
	} `json:"term"`
}

type ElasticCancelQuery struct {
	Term struct {
		Cancel bool `json:"data.Cancel"`
	} `json:"term"`
}

type ElasticDelieveredQuery struct {
	Term struct {
		Delievered bool `json:"data.Delievered"`
	} `json:"term"`
}

type ElasticFixedCountryNameQuery struct {
	Term struct {
		FixedCountryName string `json:"data.fixedCountryName"`
	} `json:"term"`
}

type ElasticFixedStateNameQuery struct {
	Term struct {
		FixedStateName string `json:"data.fixedStateName"`
	} `json:"term"`
}

type ElasticMatchFixedCountryNameQuery struct {
	Term struct {
		FixedCountryName string `json:"data.fixedCountryName"`
	} `json:"match"`
}

type ElasticMatchFixedStateNameQuery struct {
	Term struct {
		FixedStateName string `json:"data.fixedStateName"`
	} `json:"match"`
}

type ElasticBoolShouldQuery struct {
	Bool struct {
		Should []interface{} `json:"should"`
	} `json:"bool"`
}

type ElasticClientQuery struct {
	Term struct {
		Client string `json:"data.Client"`
	} `json:"match"`
}

type ElasticStateNameMatchQuery struct {
	Match struct {
		StateName string `json:"data.stateName"`
	} `json:"match"`
}

type ElasticCityNameMatchQuery struct {
	Match struct {
		CityName string `json:"data.cityName"`
	} `json:"match"`
}

type ElasticTagQuery struct {
	Term struct {
		Tag string `json:"data.Tags"`
	} `json:"match"`
}

type ElasticWritingInformationBeatsQuery struct {
	Term struct {
		Beats string `json:"data.writingInformation.beats"`
	} `json:"match"`
}

type ElasticEmployersQuery struct {
	Term struct {
		Employers string `json:"data.Employers"`
	} `json:"match"`
}

type ElasticAllQuery struct {
	Term struct {
		All string `json:"_all"`
	} `json:"match"`
}

type ElasticArchivedQuery struct {
	Term struct {
		Archived bool `json:"data.Archived"`
	} `json:"term"`
}

type ElasticIsFreelancerQuery struct {
	Term struct {
		IsFreelancer bool `json:"data.writingInformation.isFreelancer"`
	} `json:"term"`
}

type ElasticIsInfluencerQuery struct {
	Term struct {
		IsInfluencer bool `json:"data.writingInformation.isInfluencer"`
	} `json:"term"`
}

type ElasticListIdQuery struct {
	Term struct {
		ListId int64 `json:"data.ListId"`
	} `json:"term"`
}

type ElasticEmailIdQuery struct {
	Term struct {
		EmailId int64 `json:"data.EmailId"`
	} `json:"term"`
}

type ElasticUserIdQuery struct {
	Term struct {
		UserId int64 `json:"data.UserId"`
	} `json:"term"`
}

type ElasticEmailToQuery struct {
	Term struct {
		To string `json:"data.To"`
	} `json:"term"`
}

type ElasticContactIdQuery struct {
	Term struct {
		ContactId int64 `json:"data.ContactId"`
	} `json:"term"`
}

type ElasticUsernameQuery struct {
	Term struct {
		Username string `json:"data.Username"`
	} `json:"term"`
}

type ElasticInstagramUsernameQuery struct {
	Term struct {
		InstagramUsername string `json:"data.InstagramUsername"`
	} `json:"term"`
}

type ElasticFeedUsernameQuery struct {
	Term struct {
		Type     string `json:"data.Type"`
		Username string `json:"data.Username"`
	} `json:"term"`
}

type ElasticUsernameMatchQuery struct {
	Match struct {
		Username string `json:"data.Username"`
	} `json:"match"`
}

type ElasticPublicationIdQuery struct {
	Term struct {
		PublicationId int64 `json:"data.PublicationId"`
	} `json:"term"`
}

type ElasticFeedUrlQuery struct {
	Match struct {
		FeedURL string `json:"data.FeedURL"`
	} `json:"match"`
}

type ElasticBounceQuery struct {
	Term struct {
		BaseBounced bool `json:"data.Bounced"`
	} `json:"term"`
}

type ElasticCreatedRangeQuery struct {
	Range struct {
		DataCreated struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"data.Created"`
	} `json:"range"`
}

type ElasticOpenedRangeQuery struct {
	Range struct {
		DataOpened struct {
			GTE int64 `json:"gte"`
		} `json:"data.Opened"`
	} `json:"range"`
}

type ElasticClickedRangeQuery struct {
	Range struct {
		DataClicked struct {
			GTE int64 `json:"gte"`
		} `json:"data.Clicked"`
	} `json:"range"`
}

type ElasticSortDataPublishDateQuery struct {
	DataPublishDate struct {
		Order string `json:"order"`
		Mode  string `json:"mode"`
	} `json:"data.PublishDate"`
}

type ElasticSortDataCreatedAtQuery struct {
	DataCreatedAt struct {
		Order string `json:"order"`
		Mode  string `json:"mode"`
	} `json:"data.CreatedAt"`
}

type ElasticSortDataCreatedLowerQuery struct {
	DataCreated struct {
		Order string `json:"order"`
		Mode  string `json:"mode"`
	} `json:"data.created"`
}

type ElasticSortDataCreatedQuery struct {
	DataCreated struct {
		Order string `json:"order"`
		Mode  string `json:"mode"`
	} `json:"data.Created"`
}

type ElasticSortDataQuery struct {
	Date struct {
		Order string `json:"order"`
		Mode  string `json:"mode"`
	} `json:"data.Date"`
}

func InitializeElasticSearch() {
	tweetElastic := elastic.Elastic{}
	tweetElastic.BaseURL = NewBaseURL
	tweetElastic.Index = "tweets"
	tweetElastic.Type = "tweet,md-tweet"
	elasticTweet = &tweetElastic

	twitterUserElastic := elastic.Elastic{}
	twitterUserElastic.BaseURL = NewBaseURL
	twitterUserElastic.Index = "tweets"
	twitterUserElastic.Type = "user"
	elasticTwitterUser = &twitterUserElastic

	contactDatabaseElastic := elastic.Elastic{}
	contactDatabaseElastic.BaseURL = NewBaseURL
	contactDatabaseElastic.Index = "database"
	contactDatabaseElastic.Type = "contacts"
	elasticContactDatabase = &contactDatabaseElastic

	locationCountryElastic := elastic.Elastic{}
	locationCountryElastic.BaseURL = NewBaseURL
	locationCountryElastic.Index = "locations"
	locationCountryElastic.Type = "country"
	elasticLocationCountry = &locationCountryElastic

	locationStateElastic := elastic.Elastic{}
	locationStateElastic.BaseURL = NewBaseURL
	locationStateElastic.Index = "locations"
	locationStateElastic.Type = "state"
	elasticLocationState = &locationStateElastic

	locationCityElastic := elastic.Elastic{}
	locationCityElastic.BaseURL = NewBaseURL
	locationCityElastic.Index = "locations"
	locationCityElastic.Type = "city"
	elasticLocationCity = &locationCityElastic

	mediaDatabaseElastic := elastic.Elastic{}
	mediaDatabaseElastic.BaseURL = NewBaseURL
	mediaDatabaseElastic.Index = "md1"
	mediaDatabaseElastic.Type = "contacts"
	elasticMediaDatabase = &mediaDatabaseElastic

	mediaDatabasePublicationElastic := elastic.Elastic{}
	mediaDatabasePublicationElastic.BaseURL = NewBaseURL
	mediaDatabasePublicationElastic.Index = "md1"
	mediaDatabasePublicationElastic.Type = "publications"
	elasticMediaDatabasePublication = &mediaDatabasePublicationElastic

	headlineElastic := elastic.Elastic{}
	headlineElastic.BaseURL = NewBaseURL
	headlineElastic.Index = "headlines"
	headlineElastic.Type = "headline"
	elasticHeadline = &headlineElastic

	feedElastic := elastic.Elastic{}
	feedElastic.BaseURL = NewBaseURL
	feedElastic.Index = "feeds"
	feedElastic.Type = "feed,md-feed"
	elasticFeed = &feedElastic

	instagramElastic := elastic.Elastic{}
	instagramElastic.BaseURL = NewBaseURL
	instagramElastic.Index = "instagrams"
	instagramElastic.Type = "instagram"
	elasticInstagram = &instagramElastic

	instagramUserElastic := elastic.Elastic{}
	instagramUserElastic.BaseURL = NewBaseURL
	instagramUserElastic.Index = "instagrams"
	instagramUserElastic.Type = "user"
	elasticInstagramUser = &instagramUserElastic

	instagramTimeseriesElastic := elastic.Elastic{}
	instagramTimeseriesElastic.BaseURL = NewBaseURL
	instagramTimeseriesElastic.Index = "timeseries"
	instagramTimeseriesElastic.Type = "instagram"
	elasticInstagramTimeseries = &instagramTimeseriesElastic

	twitterTimeseriesElastic := elastic.Elastic{}
	twitterTimeseriesElastic.BaseURL = NewBaseURL
	twitterTimeseriesElastic.Index = "timeseries"
	twitterTimeseriesElastic.Type = "twitter"
	elasticTwitterTimeseries = &twitterTimeseriesElastic
}
