package controllers

import (
	"net/http"
	"strings"

	"google.golang.org/appengine/datastore"

	gcontext "github.com/gorilla/context"
)

// At some point automate this
var normalized = map[string]string{
	"createdby":          "CreatedBy",
	"firstname":          "FirstName",
	"lastname":           "LastName",
	"pastemployers":      "PastEmployers",
	"muckrack":           "MuckRack",
	"customfields":       "CustomFields",
	"ismastercontact":    "IsMasterContact",
	"parentcontact":      "ParentContact",
	"isoutdated":         "IsOutdated",
	"linkedinupdated":    "LinkedInUpdated",
	"sendgridid":         "SendGridId",
	"bouncedreason":      "BouncedReason",
	"issent":             "IsSent",
	"filename":           "FileName",
	"listid":             "ListId",
	"fileexists":         "FileExists",
	"fieldsmap":          "FieldsMap",
	"noticationid":       "NoticationId",
	"objectid":           "ObjectId",
	"noticationobjectid": "NoticationObjectId",
	"canwrite":           "CanWrite",
	"userid":             "UserId",
	"googleid":           "GoogleId",
	"apikey":             "ApiKey",
	"emailconfirmed":     "EmailConfirmed",
	"sendat":             "SendAt",
}

func normalizeOrderQuery(order string) string {
	operator := ""
	if string(order[0]) == "-" {
		operator = string(order[0])
		order = order[1:]
	}

	order = strings.ToLower(order)

	// If it is inside the abnormal cases above
	if normalizedOrder, ok := normalized[order]; ok {
		return operator + normalizedOrder
	}

	// Else return the titled version of it
	order = strings.Title(order)
	return operator + order
}

func ConstructQuery(query *datastore.Query, r *http.Request) *datastore.Query {
	order := gcontext.Get(r, "order").(string)
	offset := gcontext.Get(r, "offset").(int)
	limit := gcontext.Get(r, "limit").(int)
	after := gcontext.Get(r, "after").(string)

	query = query.Limit(limit)

	if order != "" {
		query = query.Order(normalizeOrderQuery(order))
	}

	if after != "" {
		cursor, err := datastore.DecodeCursor(after)
		if err == nil {
			query = query.Start(cursor)
		}
	} else {
		query = query.Offset(offset)
	}

	return query
}
