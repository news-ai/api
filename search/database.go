package search

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/context"

	gcontext "github.com/gorilla/context"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	apiModels "github.com/news-ai/api/models"
	pitchModels "github.com/news-ai/pitch/models"

	elastic "github.com/news-ai/elastic-appengine"
)

var (
	elasticLocationCity             *elastic.Elastic
	elasticLocationState            *elastic.Elastic
	elasticLocationCountry          *elastic.Elastic
	elasticContactDatabase          *elastic.Elastic
	elasticMediaDatabase            *elastic.Elastic
	elasticMediaDatabasePublication *elastic.Elastic
)

type EnhanceResponse struct {
	Data interface{} `json:"data"`
}

type EnhanceFullContactProfileVerifyResponse struct {
	Data struct {
		Enrich EnhanceFullContactProfileResponse `json:"enrich"`
		Verify EnhanceFullContactVerifyResponse  `json:"verify"`
	} `json:"data"`
}

type EnhanceFullContactVerifyResponse struct {
	Data struct {
		RequestID string `json:"requestId"`
		Status    int    `json:"status"`
		Email     struct {
			Message    string `json:"message"`
			Person     string `json:"person"`
			Username   string `json:"username"`
			SendSafely bool   `json:"sendSafely"`
			Corrected  bool   `json:"corrected"`
			Address    string `json:"address"`
			Company    string `json:"company"`
			Domain     string `json:"domain"`
			Attributes struct {
				Disposable  bool `json:"disposable"`
				Catchall    bool `json:"catchall"`
				Risky       bool `json:"risky"`
				Deliverable bool `json:"deliverable"`
				ValidSyntax bool `json:"validSyntax"`
			} `json:"attributes"`
		} `json:"email"`
	} `json:"data"`
}

type EnhanceFullContactProfileResponse struct {
	Data struct {
		Status int `json:"status"`

		Organizations    []pitchModels.Organization   `json:"organizations"`
		DigitalFootprint pitchModels.DigitalFootprint `json:"digitalFootprint"`
		SocialProfiles   []pitchModels.SocialProfile  `json:"socialProfiles"`
		Demographics     pitchModels.Demographic      `json:"demographics"`
		Photos           []pitchModels.Photo          `json:"photos"`
		ContactInfo      pitchModels.ContactInfo      `json:"contactInfo"`

		RequestID  string  `json:"requestId"`
		Likelihood float64 `json:"likelihood"`
	} `json:"data"`
}

type EnhanceFullContactCompanyResponse struct {
	Data struct {
		Status    int    `json:"status"`
		RequestID string `json:"requestId"`
		Category  []struct {
			Name string `json:"name"`
			Code string `json:"code"`
		} `json:"category"`
		Logo           string `json:"logo"`
		Website        string `json:"website"`
		LanguageLocale string `json:"languageLocale"`
		OnlineSince    string `json:"onlineSince"`
		Organization   struct {
			Name            string `json:"name"`
			ApproxEmployees int    `json:"approxEmployees"`
			Founded         string `json:"founded"`
			ContactInfo     struct {
				EmailAddresses []struct {
					Value string `json:"value"`
					Label string `json:"label"`
				} `json:"emailAddresses"`
				PhoneNumbers []struct {
					Number string `json:"number"`
					Label  string `json:"label"`
				} `json:"phoneNumbers"`
				Addresses []struct {
					AddressLine1 string `json:"addressLine1"`
					Locality     string `json:"locality"`
					Region       struct {
						Name string `json:"name"`
						Code string `json:"code"`
					} `json:"region"`
					Country struct {
						Name string `json:"name"`
						Code string `json:"code"`
					} `json:"country"`
					PostalCode string `json:"postalCode"`
					Label      string `json:"label"`
				} `json:"addresses"`
			} `json:"contactInfo"`
			Links []struct {
				URL   string `json:"url"`
				Label string `json:"label"`
			} `json:"links"`
			Images []struct {
				URL   string `json:"url"`
				Label string `json:"label"`
			} `json:"images"`
			Keywords []string `json:"keywords"`
		} `json:"organization"`
		SocialProfiles []pitchModels.SocialProfile `json:"socialProfiles"`
	} `json:"data"`
}

type EnhanceEmailVerificationResponse struct {
	Data struct {
		Status    int    `json:"status"`
		RequestID string `json:"requestId"`
		Emails    []struct {
			Message    string `json:"message"`
			Address    string `json:"address"`
			Username   string `json:"username"`
			Domain     string `json:"domain"`
			Corrected  bool   `json:"corrected"`
			Attributes struct {
				ValidSyntax bool `json:"validSyntax"`
				Deliverable bool `json:"deliverable"`
				Catchall    bool `json:"catchall"`
				Risky       bool `json:"risky"`
				Disposable  bool `json:"disposable"`
			} `json:"attributes"`
			Person     string `json:"person"`
			Company    string `json:"company"`
			SendSafely bool   `json:"sendSafely"`
		} `json:"emails"`
	} `json:"data"`
}

type SearchMediaDatabaseInner struct {
	Beats           []string `json:"beats"`
	OccasionalBeats []string `json:"occasionalBeats"`

	// Single-option fields (so no AND or OR)
	IsFreelancer bool `json:"isFreelancer"`
	IsInfluencer bool `json:"isInfluencer"`

	// Could be both AND or OR. Works for X and Y or all contacts that
	// work for X or Y
	Organizations []string `json:"organizations"`

	// Locations is definitely an OR field
	Locations []struct {
		Country string `json:"country"`
		State   string `json:"state"`
		City    string `json:"city"`
	} `json:"locations"`

	// Search RSS-related fields
	RSS struct {
		Headline    string `json:"headline"`
		IncludeBody bool   `json:"includeBody"`
	} `json:"rss"`

	// Search Instagram-related fields
	Instagram struct {
		Description string `json:"description"`
	} `json:"instagram"`

	// Search Twitter-related fields
	Twitter struct {
		TweetBody       string `json:"tweetbody"`
		UserDescription string `json:"userDescription"`
	} `json:"twitter"`

	Time struct {
		From time.Time `json:"from"`
		To   time.Time `json:"to"`
	} `json:"time"`
}

type SearchMediaDatabaseQuery struct {
	Included SearchMediaDatabaseInner `json:"included"`
	Excluded SearchMediaDatabaseInner `json:"excluded"`
}

type DatabaseResponse struct {
	Email string      `json:"email"`
	Data  interface{} `json:"data"`
}

type LocationCityResponse struct {
	Id               string `json:"id"`
	FixedCountryName string `json:"countryName"`
	FixedStateName   string `json:"stateName"`
	CityName         string `json:"cityName"`
}

func (lcr *LocationCityResponse) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := apiModels.SetField(lcr, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

type LocationStateResponse struct {
	Id               string `json:"id"`
	FixedCountryName string `json:"countryName"`
	StateName        string `json:"stateName"`
}

func (lsr *LocationStateResponse) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := apiModels.SetField(lsr, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

type LocationCountryResponse struct {
	Id          string `json:"id"`
	CountryName string `json:"countryName"`
}

func (lcr *LocationCountryResponse) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := apiModels.SetField(lcr, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchESMediaDatabase(c context.Context, elasticQuery interface{}) (interface{}, int, int, error) {
	hits, err := elasticMediaDatabase.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, 0, err
	}

	contactHits := hits.Hits
	if len(contactHits) == 0 {
		log.Infof(c, "%v", hits)
		return nil, 0, 0, nil
	}

	var interfaceSlice = make([]interface{}, len(contactHits))
	for i := 0; i < len(contactHits); i++ {
		interfaceSlice[i] = contactHits[i].Source.Data
	}

	return interfaceSlice, len(contactHits), hits.Total, nil
}

func searchESMediaDatabasePublication(c context.Context, elasticQuery interface{}) (interface{}, int, int, error) {
	hits, err := elasticMediaDatabasePublication.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, 0, err
	}

	publicationHits := hits.Hits
	if len(publicationHits) == 0 {
		log.Infof(c, "%v", hits)
		return nil, 0, 0, errors.New("No media database publications")
	}

	publications := []pitchModels.Publication{}
	for i := 0; i < len(publicationHits); i++ {
		rawPublication := publicationHits[i].Source.Data
		rawMap := rawPublication.(map[string]interface{})
		publication := pitchModels.Publication{}
		err := publication.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		publication.Id = publicationHits[i].ID
		publication.Type = "media-publications"
		publications = append(publications, publication)
	}

	return publications, len(publications), hits.Total, nil
}

func searchESContactsDatabase(c context.Context, elasticQuery elastic.ElasticQuery) (interface{}, int, int, error) {
	hits, err := elasticContactDatabase.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, 0, err
	}

	contactHits := hits.Hits
	var contacts []interface{}
	for i := 0; i < len(contactHits); i++ {
		rawContact := contactHits[i].Source.Data
		contactData := DatabaseResponse{
			Email: contactHits[i].ID,
			Data:  rawContact,
		}
		contacts = append(contacts, contactData)
	}

	return contacts, len(contactHits), hits.Total, nil
}

func SearchEnhanceForEmailVerification(c context.Context, r *http.Request, email string) (EnhanceEmailVerificationResponse, error) {
	contextWithTimeout, _ := context.WithTimeout(c, time.Second*15)
	client := urlfetch.Client(contextWithTimeout)
	getUrl := "https://enhance.newsai.org/verify/" + email

	req, _ := http.NewRequest("GET", getUrl, nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", err)
		return EnhanceEmailVerificationResponse{}, err
	}
	defer resp.Body.Close()

	var enhanceResponse EnhanceEmailVerificationResponse
	err = json.NewDecoder(resp.Body).Decode(&enhanceResponse)
	if err != nil {
		log.Errorf(c, "%v", err)
		return EnhanceEmailVerificationResponse{}, err
	}

	return enhanceResponse, nil
}
func SearchCompanyDatabase(c context.Context, r *http.Request, url string) (EnhanceFullContactCompanyResponse, error) {
	contextWithTimeout, _ := context.WithTimeout(c, time.Second*15)
	client := urlfetch.Client(contextWithTimeout)
	getUrl := "https://enhance.newsai.org/company/" + url

	req, _ := http.NewRequest("GET", getUrl, nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", err)
		return EnhanceFullContactCompanyResponse{}, err
	}
	defer resp.Body.Close()

	var enhanceResponse EnhanceFullContactCompanyResponse
	err = json.NewDecoder(resp.Body).Decode(&enhanceResponse)
	if err != nil {
		log.Errorf(c, "%v", err)
		return EnhanceFullContactCompanyResponse{}, err
	}

	return enhanceResponse, nil
}

func SearchContactDatabase(c context.Context, r *http.Request, email string) (EnhanceFullContactProfileResponse, error) {
	contextWithTimeout, _ := context.WithTimeout(c, time.Second*8)
	client := urlfetch.Client(contextWithTimeout)
	getUrl := "https://enhance.newsai.org/fullcontact/" + email

	req, _ := http.NewRequest("GET", getUrl, nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", err)
		return EnhanceFullContactProfileResponse{}, err
	}
	defer resp.Body.Close()

	var enhanceResponse EnhanceFullContactProfileResponse
	err = json.NewDecoder(resp.Body).Decode(&enhanceResponse)
	if err != nil {
		log.Errorf(c, "%v", err)
		return EnhanceFullContactProfileResponse{}, err
	}

	return enhanceResponse, nil
}

func SearchContactVerifyDatabase(c context.Context, r *http.Request, email string) (EnhanceFullContactProfileVerifyResponse, error) {
	contextWithTimeout, _ := context.WithTimeout(c, time.Second*8)
	client := urlfetch.Client(contextWithTimeout)
	getUrl := "https://enhance.newsai.org/fullcontact2/" + email

	req, _ := http.NewRequest("GET", getUrl, nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", err)
		return EnhanceFullContactProfileVerifyResponse{}, err
	}
	defer resp.Body.Close()

	var enhanceResponse EnhanceFullContactProfileVerifyResponse
	err = json.NewDecoder(resp.Body).Decode(&enhanceResponse)
	if err != nil {
		log.Errorf(c, "%v", err)
		return EnhanceFullContactProfileVerifyResponse{}, err
	}

	return enhanceResponse, nil
}

func SearchContactDatabaseForMediaDatbase(c context.Context, r *http.Request, email string) (pitchModels.MediaDatabaseProfile, error) {
	contextWithTimeout, _ := context.WithTimeout(c, time.Second*15)
	client := urlfetch.Client(contextWithTimeout)
	getUrl := "https://enhance.newsai.org/fullcontact/" + email

	req, _ := http.NewRequest("GET", getUrl, nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", err)
		return pitchModels.MediaDatabaseProfile{}, err
	}
	defer resp.Body.Close()

	var enhanceResponse pitchModels.MediaDatabaseProfile
	err = json.NewDecoder(resp.Body).Decode(&enhanceResponse)
	if err != nil {
		log.Errorf(c, "%v", err)
		return pitchModels.MediaDatabaseProfile{}, err
	}

	return enhanceResponse, nil
}

func SearchContactInMediaDatabase(c context.Context, r *http.Request, email string) (pitchModels.MediaDatabaseProfile, error) {
	contextWithTimeout, _ := context.WithTimeout(c, time.Second*15)
	client := urlfetch.Client(contextWithTimeout)
	getUrl := "https://enhance.newsai.org/md/" + email

	req, _ := http.NewRequest("GET", getUrl, nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf(c, "%v", err)
		return pitchModels.MediaDatabaseProfile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New("Invalid response from ES")
		log.Infof(c, "%v", err)
		return pitchModels.MediaDatabaseProfile{}, err
	}

	var enhanceResponse pitchModels.MediaDatabaseProfile
	err = json.NewDecoder(resp.Body).Decode(&enhanceResponse)
	if err != nil {
		log.Errorf(c, "%v", err)
		return pitchModels.MediaDatabaseProfile{}, err
	}

	if enhanceResponse.Data.Status != 200 {
		err = errors.New("Could not retrieve profile from ES")
		log.Infof(c, "%v", err)
		return pitchModels.MediaDatabaseProfile{}, err
	}

	return enhanceResponse, nil
}

func SearchESMediaDatabasePublications(c context.Context, r *http.Request) (interface{}, int, int, error) {
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedQuery := ElasticSortDataCreatedLowerQuery{}
	elasticCreatedQuery.DataCreated.Order = "desc"
	elasticCreatedQuery.DataCreated.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedQuery)

	return searchESMediaDatabasePublication(c, elasticQuery)
}

func SearchESMediaDatabase(c context.Context, r *http.Request) (interface{}, int, int, error) {
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryWithSort{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedQuery := ElasticSortDataCreatedLowerQuery{}
	elasticCreatedQuery.DataCreated.Order = "desc"
	elasticCreatedQuery.DataCreated.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedQuery)

	return searchESMediaDatabase(c, elasticQuery)
}

func SearchContactsInESMediaDatabase(c context.Context, r *http.Request, searchQuery SearchMediaDatabaseQuery) (interface{}, int, int, error) {
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQueryWithSortShould{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticCreatedQuery := ElasticSortDataCreatedLowerQuery{}
	elasticCreatedQuery.DataCreated.Order = "desc"
	elasticCreatedQuery.DataCreated.Mode = "avg"
	elasticQuery.Sort = append(elasticQuery.Sort, elasticCreatedQuery)

	if len(searchQuery.Included.Organizations) == 1 {
		if searchQuery.Included.Organizations[0] != "" {
			elasticOrganizationNameQuery := ElasticOrganizationNameQuery{}
			elasticOrganizationNameQuery.Match.Name = searchQuery.Included.Organizations[0]
			elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticOrganizationNameQuery)
		}
	} else if len(searchQuery.Included.Organizations) > 1 {
		for i := 0; i < len(searchQuery.Included.Organizations); i++ {
			elasticOrganizationNameQuery := ElasticOrganizationNameQuery{}
			elasticOrganizationNameQuery.Match.Name = searchQuery.Included.Organizations[i]
			elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticOrganizationNameQuery)
		}
	}

	if len(searchQuery.Included.Locations) == 1 {
		if searchQuery.Included.Locations[0].City != "" {
			elasticLocationCityQuery := ElasticLocationCityQuery{}
			elasticLocationCityQuery.Term.City = searchQuery.Included.Locations[0].City
			elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticLocationCityQuery)
		}

		if searchQuery.Included.Locations[0].State != "" {
			elasticLocationStateQuery := ElasticLocationStateQuery{}
			elasticLocationStateQuery.Term.State = searchQuery.Included.Locations[0].State
			elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticLocationStateQuery)
		}

		if searchQuery.Included.Locations[0].Country != "" {
			elasticLocationCountryQuery := ElasticLocationCountryQuery{}
			elasticLocationCountryQuery.Term.Country = searchQuery.Included.Locations[0].Country
			elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticLocationCountryQuery)
		}
	} else if len(searchQuery.Included.Locations) > 1 {
		for i := 0; i < len(searchQuery.Included.Locations); i++ {
			// We do a "should" query on multiple locations. But, we only
			// filter by cities. If we filter by states then it would give us
			// all of the
			elasticLocationCityQuery := ElasticLocationCityQuery{}
			elasticLocationCityQuery.Term.City = searchQuery.Included.Locations[i].City
			elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticLocationCityQuery)
		}
	}

	if len(searchQuery.Included.Beats) == 1 {
		elasticBeatsQuery := ElasticWritingInformationBeatsQuery{}
		elasticBeatsQuery.Term.Beats = searchQuery.Included.Beats[0]
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticBeatsQuery)
	} else if len(searchQuery.Included.Beats) > 1 {
		for i := 0; i < len(searchQuery.Included.Beats); i++ {
			elasticBeatsQuery := ElasticWritingInformationBeatsQuery{}
			elasticBeatsQuery.Term.Beats = searchQuery.Included.Beats[i]
			elasticQuery.Query.Bool.Should = append(elasticQuery.Query.Bool.Should, elasticBeatsQuery)
		}
	}

	if searchQuery.Included.IsFreelancer {
		elasticIsFreelancerQuery := ElasticIsFreelancerQuery{}
		elasticIsFreelancerQuery.Term.IsFreelancer = searchQuery.Included.IsFreelancer
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticIsFreelancerQuery)
	}

	if searchQuery.Included.IsInfluencer {
		elasticIsInfluencerQuery := ElasticIsInfluencerQuery{}
		elasticIsInfluencerQuery.Term.IsInfluencer = searchQuery.Included.IsInfluencer
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticIsInfluencerQuery)
	}

	return searchESMediaDatabase(c, elasticQuery)
}

func SearchESContactsDatabase(c context.Context, r *http.Request) (interface{}, int, int, error) {
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	return searchESContactsDatabase(c, elasticQuery)
}

func ESCityLocation(c context.Context, r *http.Request, cityName, stateName, countryName string) (interface{}, int, int, error) {
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticMatchFixedCountryNameQuery := ElasticMatchFixedCountryNameQuery{}
	elasticMatchFixedCountryNameQuery.Term.FixedCountryName = countryName
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticMatchFixedCountryNameQuery)

	elasticMatchFixedStateNameQuery := ElasticMatchFixedStateNameQuery{}
	elasticMatchFixedStateNameQuery.Term.FixedStateName = stateName
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticMatchFixedStateNameQuery)

	cityName = strings.Replace(cityName, "\"", "", -1)
	if cityName != "" {
		elasticCityNameMatchQuery := ElasticCityNameMatchQuery{}
		elasticCityNameMatchQuery.Match.CityName = cityName

		elasticBoolShouldQuery := ElasticBoolShouldQuery{}
		elasticBoolShouldQuery.Bool.Should = append(elasticBoolShouldQuery.Bool.Should, elasticCityNameMatchQuery)
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticBoolShouldQuery)
	}

	hits, err := elasticLocationCity.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, 0, err
	}

	locationHits := hits.Hits
	if len(locationHits) == 0 {
		log.Infof(c, "%v", hits)
		return nil, 0, 0, nil
	}

	cities := []LocationCityResponse{}
	for i := 0; i < len(locationHits); i++ {
		rawCity := locationHits[i].Source.Data
		rawMap := rawCity.(map[string]interface{})
		city := LocationCityResponse{}
		err := city.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		city.Id = locationHits[i].ID

		if len(locationHits) > 1 && cityName != "" {
			lowerCaseCityName := strings.ToLower(city.CityName)
			lowerCaseFixedCityName := strings.ToLower(cityName)
			if lowerCaseCityName[0] == lowerCaseFixedCityName[0] && strings.ToLower(city.FixedStateName) == strings.ToLower(stateName) && strings.ToLower(city.FixedCountryName) == strings.ToLower(countryName) {
				if strings.Contains(lowerCaseCityName, lowerCaseFixedCityName) {
					cities = append(cities, city)
				}
			}
		} else {
			cities = append(cities, city)
		}
	}

	if len(cities) == 0 {
		return nil, 0, 0, nil
	}

	if strings.ToLower(cityName) == strings.ToLower(cities[0].CityName) && strings.ToLower(stateName) == strings.ToLower(cities[0].FixedStateName) && strings.ToLower(countryName) == strings.ToLower(cities[0].FixedCountryName) {
		return []LocationCityResponse{cities[0]}, 1, 1, nil
	}

	return cities, len(cities), hits.Total, nil
}

func ESStateLocation(c context.Context, r *http.Request, stateName string, countryName string) (interface{}, int, int, error) {
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	elasticQuery := elastic.ElasticQuery{}
	elasticQuery.Size = limit
	elasticQuery.From = offset

	elasticMatchFixedCountryNameQuery := ElasticMatchFixedCountryNameQuery{}
	elasticMatchFixedCountryNameQuery.Term.FixedCountryName = countryName
	elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticMatchFixedCountryNameQuery)

	stateName = strings.Replace(stateName, "\"", "", -1)
	if stateName != "" {
		elasticStateNameMatchQuery := ElasticStateNameMatchQuery{}
		elasticStateNameMatchQuery.Match.StateName = stateName

		elasticBoolShouldQuery := ElasticBoolShouldQuery{}
		elasticBoolShouldQuery.Bool.Should = append(elasticBoolShouldQuery.Bool.Should, elasticStateNameMatchQuery)
		elasticQuery.Query.Bool.Must = append(elasticQuery.Query.Bool.Must, elasticBoolShouldQuery)
	}

	hits, err := elasticLocationState.QueryStruct(c, elasticQuery)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, 0, err
	}

	locationHits := hits.Hits
	if len(locationHits) == 0 {
		log.Infof(c, "%v", hits)
		return nil, 0, 0, nil
	}

	states := []LocationStateResponse{}
	for i := 0; i < len(locationHits); i++ {
		rawState := locationHits[i].Source.Data
		rawMap := rawState.(map[string]interface{})
		state := LocationStateResponse{}
		err := state.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		state.Id = locationHits[i].ID

		if len(locationHits) > 1 && stateName != "" {
			lowerCaseStateName := strings.ToLower(state.StateName)
			lowerCaseFixedStateName := strings.ToLower(stateName)
			if lowerCaseStateName[0] == lowerCaseFixedStateName[0] && strings.ToLower(state.FixedCountryName) == strings.ToLower(countryName) {
				if strings.Contains(lowerCaseStateName, lowerCaseFixedStateName) {
					states = append(states, state)
				}
			}
		} else {
			states = append(states, state)
		}
	}

	if len(states) == 0 {
		return nil, 0, 0, nil
	}

	if strings.ToLower(stateName) == strings.ToLower(states[0].StateName) && strings.ToLower(countryName) == strings.ToLower(states[0].FixedCountryName) {
		return []LocationStateResponse{states[0]}, 1, 1, nil
	}

	return states, len(states), hits.Total, nil
}

func ESCountryLocation(c context.Context, r *http.Request, countryName string) (interface{}, int, int, error) {
	if countryName == "" {
		return nil, 0, 0, nil
	}

	countryName = strings.Replace(countryName, "\"", "", -1)
	search := ""
	if countryName != "" {
		search = url.QueryEscape(countryName)
		search = "q=data.countryName:" + search
	}

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	hits, err := elasticLocationCountry.Query(c, offset, limit, search)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, 0, 0, err
	}

	locationHits := hits.Hits
	if len(locationHits) == 0 {
		log.Infof(c, "%v", hits)
		return nil, 0, 0, nil
	}

	countries := []LocationCountryResponse{}
	for i := 0; i < len(locationHits); i++ {
		rawCountry := locationHits[i].Source.Data
		rawMap := rawCountry.(map[string]interface{})
		country := LocationCountryResponse{}
		err := country.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		country.Id = locationHits[i].ID

		// In this case we only want countries
		// starting with that letter. "United" shouldn't
		// return "Tanzania" even though it has "United"
		// in it's name
		// unless there's only 1 thing matching. Then we can
		// just allow it to pass (like Holy See, which can
		// be search with Vatican)
		if len(locationHits) > 1 && countryName != "" {
			lowerCaseCountryName := strings.ToLower(country.CountryName)
			lowerCaseFixedCountryName := strings.ToLower(countryName)
			if lowerCaseCountryName[0] == lowerCaseFixedCountryName[0] {
				// We can now filter out the ones that don't seem like they match
				if strings.Contains(lowerCaseCountryName, lowerCaseFixedCountryName) {
					countries = append(countries, country)
				}
			}
		} else {
			countries = append(countries, country)
		}
	}

	// Check again if anything matches
	if len(countries) == 0 {
		return nil, 0, 0, nil
	}

	// This means that we've found the country, just return this
	// no point in suggesting more
	if strings.ToLower(countries[0].CountryName) == strings.ToLower(countryName) {
		return []LocationCountryResponse{countries[0]}, 1, 1, nil
	}

	return countries, len(countries), hits.Total, nil
}

func SearchPublicationInESMediaDatabase(c context.Context, r *http.Request, search string) ([]pitchModels.Publication, int, error) {
	search = url.QueryEscape(search)
	search = "q=data.organizationName:" + search

	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)

	hits, err := elasticMediaDatabasePublication.Query(c, offset, limit, search)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []pitchModels.Publication{}, 0, err
	}

	publicationHits := hits.Hits
	publications := []pitchModels.Publication{}
	for i := 0; i < len(publicationHits); i++ {
		rawPublication := publicationHits[i].Source.Data
		rawMap := rawPublication.(map[string]interface{})
		publication := pitchModels.Publication{}
		err := publication.FillStruct(rawMap)
		if err != nil {
			log.Errorf(c, "%v", err)
		}

		publication.Id = publicationHits[i].ID
		publication.Type = "publications"
		publications = append(publications, publication)
	}

	return publications, hits.Total, nil
}

func GetMediaDatabaseContactsSchema(c context.Context) (interface{}, error) {
	mapping, err := elasticMediaDatabase.GetMapping(c)
	if err != nil {
		log.Errorf(c, "%v", err)
		return nil, err
	}

	return mapping, nil
}
