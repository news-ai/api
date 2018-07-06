package models

import (
	"errors"
	"reflect"

	"github.com/news-ai/cast"
)

func SetField(obj interface{}, name string, value interface{}) error {
	if name == "id" {
		name = "Id"
	}

	// Rename the lowercase organization name
	if name == "organizationName" {
		name = "OrganizationName"
	}

	if name == "countryName" {
		name = "CountryName"
	}

	if name == "fixedCountryName" {
		name = "FixedCountryName"
	}

	if name == "stateName" {
		name = "StateName"
	}

	if name == "fixedStateName" {
		name = "FixedStateName"
	}

	if name == "cityName" {
		name = "CityName"
	}

	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return errors.New("No such field:" + name + " in obj")
	}

	if !structFieldValue.CanSet() {
		return errors.New("Cannot set" + name + " field value")
	}

	if name == "TwitterId" || name == "Entities" {
		return nil
	}

	if name == "CustomFields" {
		// customFields := []CustomContactFieldElastic{}
		// switch v := value.(type) {
		// case []interface{}:
		//  for _, u := range v {
		//      customFields = append(customFields, u.(CustomContactFieldElastic))
		//  }
		// }
		// val := reflect.ValueOf(customFields)
		// structFieldValue.Set(val)
		return nil
	}

	// Cast int
	if name == "Comments" || name == "Likes" || name == "InstagramLikes" || name == "InstagramComments" || name == "StatusesCount" || name == "UtcOffset" || name == "FavouritesCount" || name == "ListedCount" || name == "FriendsCount" || name == "FollowersCount" || name == "ID" || name == "InstagramWidth" || name == "InstagramHeight" || name == "Retweets" || name == "TwitterLikes" || name == "TwitterRetweets" || name == "Followers" || name == "Following" || name == "Posts" || name == "ExpiresIn" || name == "Clicked" || name == "Opened" || name == "SendGridOpened" || name == "SendGridClicked" {
		returnValue := cast.ToInt(value)
		val := reflect.ValueOf(returnValue)
		structFieldValue.Set(val)
		return nil
	}

	// Cast string array
	if name == "Categories" || name == "Tags" || name == "CC" || name == "BCC" {
		returnValue := cast.ToStringSlice(value)
		val := reflect.ValueOf(returnValue)
		structFieldValue.Set(val)
		return nil
	}

	// Cast time
	if name == "Created" || name == "Updated" || name == "LinkedInUpdated" || name == "PublishDate" || name == "CreatedAt" || name == "SendAt" {
		returnValue, err := cast.ToTime(value)
		if err != nil {
			return err
		}
		val := reflect.ValueOf(returnValue)
		structFieldValue.Set(val)
		return nil
	}

	// CustomFields
	if name == "CustomFields" {
		val := reflect.ValueOf(value)
		structFieldValue.Set(val)
		return nil
	}

	// Int64
	if name == "Id" || name == "CreatedBy" || name == "ParentContact" || name == "ListId" || name == "ContactId" || name == "PublicationId" || name == "TweetId" || name == "FileUpload" || name == "TeamId" || name == "TemplateId" || name == "ClientId" {
		returnValue := cast.ToInt64(value)
		val := reflect.ValueOf(returnValue)
		structFieldValue.Set(val)
		return nil
	}

	// Int64 array
	if name == "Administrators" || name == "Employers" || name == "PastEmployers" || name == "Attachments" {
		returnValue, err := cast.ToInt64SliceE(value)
		if err != nil {
			return err
		}
		val := reflect.ValueOf(returnValue)
		structFieldValue.Set(val)
		return nil
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)

	if structFieldType != val.Type() {
		invalidTypeError := errors.New("Provided value type didn't match obj field type")
		return invalidTypeError
	}

	structFieldValue.Set(val)
	return nil
}
