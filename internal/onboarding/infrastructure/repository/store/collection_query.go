package repository

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tuannm99/podzone/pkg/collection"
)

type requestFieldKind uint8

const (
	requestFieldString requestFieldKind = iota
	requestFieldObjectID
	requestFieldTime
)

type requestField struct {
	path string
	kind requestFieldKind
}

var requestCollectionFields = map[string]requestField{
	"id":          {path: "_id", kind: requestFieldObjectID},
	"workspaceId": {path: "workspace_id", kind: requestFieldString},
	"name":        {path: "name", kind: requestFieldString},
	"subdomain":   {path: "subdomain", kind: requestFieldString},
	"requestedBy": {path: "requested_by", kind: requestFieldString},
	"status":      {path: "status", kind: requestFieldString},
	"storeId":     {path: "store_id", kind: requestFieldObjectID},
	"createdAt":   {path: "created_at", kind: requestFieldTime},
	"updatedAt":   {path: "updated_at", kind: requestFieldTime},
}

var transitionCollectionFields = map[string]requestField{
	"id":        {path: "_id", kind: requestFieldObjectID},
	"requestId": {path: "request_id", kind: requestFieldString},
	"from":      {path: "from", kind: requestFieldString},
	"to":        {path: "to", kind: requestFieldString},
	"step":      {path: "step", kind: requestFieldString},
	"reason":    {path: "reason", kind: requestFieldString},
	"errorCode": {path: "error_code", kind: requestFieldString},
	"createdAt": {path: "created_at", kind: requestFieldTime},
}

func buildStoreRequestCollection(
	workspaceID string,
	query collection.Query,
) (collection.Query, bson.M, bson.D, error) {
	normalized := query.Normalize()
	clauses := bson.A{bson.M{"workspace_id": workspaceID}}
	if search := strings.TrimSpace(normalized.Search); search != "" {
		pattern := primitive.Regex{Pattern: regexp.QuoteMeta(search), Options: "i"}
		searchClauses := bson.A{
			bson.M{"name": pattern},
			bson.M{"subdomain": pattern},
			bson.M{"requested_by": pattern},
			bson.M{"status": pattern},
		}
		if id, err := primitive.ObjectIDFromHex(search); err == nil {
			searchClauses = append(searchClauses, bson.M{"_id": id})
		}
		clauses = append(clauses, bson.M{"$or": searchClauses})
	}
	for _, filter := range normalized.Filters {
		clause, err := storeRequestFilter(filter)
		if err != nil {
			return collection.Query{}, nil, nil, err
		}
		clauses = append(clauses, clause)
	}

	sortField := requestCollectionFields["updatedAt"]
	if normalized.SortBy != "" {
		var ok bool
		sortField, ok = requestCollectionFields[normalized.SortBy]
		if !ok {
			return collection.Query{}, nil, nil, fmt.Errorf(
				"%w: unsupported store request sort field %q",
				collection.ErrInvalidQuery,
				normalized.SortBy,
			)
		}
	}
	direction := -1
	if normalized.SortDirection == collection.SortAscending {
		direction = 1
	}
	return normalized, bson.M{"$and": clauses}, bson.D{
		{Key: sortField.path, Value: direction},
		{Key: "_id", Value: 1},
	}, nil
}

func buildStoreTransitionCollection(
	requestID string,
	query collection.Query,
) (collection.Query, bson.M, bson.D, error) {
	normalized := query.Normalize()
	clauses := bson.A{bson.M{"request_id": requestID}}
	if search := strings.TrimSpace(normalized.Search); search != "" {
		pattern := primitive.Regex{Pattern: regexp.QuoteMeta(search), Options: "i"}
		clauses = append(clauses, bson.M{"$or": bson.A{
			bson.M{"from": pattern},
			bson.M{"to": pattern},
			bson.M{"step": pattern},
			bson.M{"reason": pattern},
			bson.M{"error_code": pattern},
		}})
	}
	for _, filter := range normalized.Filters {
		clause, err := storeTransitionFilter(filter)
		if err != nil {
			return collection.Query{}, nil, nil, err
		}
		clauses = append(clauses, clause)
	}
	sortField := transitionCollectionFields["createdAt"]
	if normalized.SortBy != "" {
		var ok bool
		sortField, ok = transitionCollectionFields[normalized.SortBy]
		if !ok {
			return collection.Query{}, nil, nil, fmt.Errorf(
				"%w: unsupported store transition sort field %q",
				collection.ErrInvalidQuery,
				normalized.SortBy,
			)
		}
	}
	direction := -1
	if normalized.SortDirection == collection.SortAscending {
		direction = 1
	}
	return normalized, bson.M{"$and": clauses}, bson.D{
		{Key: sortField.path, Value: direction},
		{Key: "_id", Value: 1},
	}, nil
}

func storeRequestFilter(filter collection.Filter) (bson.M, error) {
	field, ok := requestCollectionFields[filter.Field]
	if !ok {
		return nil, fmt.Errorf(
			"%w: unsupported store request filter field %q",
			collection.ErrInvalidQuery,
			filter.Field,
		)
	}
	return storeCollectionFilter(field, filter)
}

func storeTransitionFilter(filter collection.Filter) (bson.M, error) {
	field, ok := transitionCollectionFields[filter.Field]
	if !ok {
		return nil, fmt.Errorf(
			"%w: unsupported store transition filter field %q",
			collection.ErrInvalidQuery,
			filter.Field,
		)
	}
	return storeCollectionFilter(field, filter)
}

func storeCollectionFilter(field requestField, filter collection.Filter) (bson.M, error) {
	if len(filter.Values) == 0 {
		return nil, fmt.Errorf("%w: filter %q requires a value", collection.ErrInvalidQuery, filter.Field)
	}
	if filter.Operator == collection.FilterContains || filter.Operator == collection.FilterStartsWith {
		if field.kind != requestFieldString {
			return nil, unsupportedStoreRequestFilter(filter)
		}
		pattern := regexp.QuoteMeta(filter.Values[0])
		if filter.Operator == collection.FilterStartsWith {
			pattern = "^" + pattern
		}
		return bson.M{field.path: primitive.Regex{Pattern: pattern, Options: "i"}}, nil
	}
	values, err := convertStoreRequestValues(field, filter.Values)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid value for field %q", collection.ErrInvalidQuery, filter.Field)
	}
	switch filter.Operator {
	case collection.FilterEqual:
		return bson.M{field.path: values[0]}, nil
	case collection.FilterNotEqual:
		return bson.M{field.path: bson.M{"$ne": values[0]}}, nil
	case collection.FilterGreaterThan:
		return bson.M{field.path: bson.M{"$gt": values[0]}}, nil
	case collection.FilterGreaterThanOrEqual:
		return bson.M{field.path: bson.M{"$gte": values[0]}}, nil
	case collection.FilterLessThan:
		return bson.M{field.path: bson.M{"$lt": values[0]}}, nil
	case collection.FilterLessThanOrEqual:
		return bson.M{field.path: bson.M{"$lte": values[0]}}, nil
	case collection.FilterIn:
		return bson.M{field.path: bson.M{"$in": values}}, nil
	default:
		return nil, unsupportedStoreRequestFilter(filter)
	}
}

func convertStoreRequestValues(field requestField, values []string) (bson.A, error) {
	out := make(bson.A, 0, len(values))
	for _, value := range values {
		switch field.kind {
		case requestFieldObjectID:
			id, err := primitive.ObjectIDFromHex(value)
			if err != nil {
				return nil, err
			}
			out = append(out, id)
		case requestFieldTime:
			timestamp, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return nil, err
			}
			out = append(out, timestamp)
		default:
			out = append(out, value)
		}
	}
	return out, nil
}

func unsupportedStoreRequestFilter(filter collection.Filter) error {
	return fmt.Errorf(
		"%w: unsupported operator %q for store request field %q",
		collection.ErrInvalidQuery,
		filter.Operator,
		filter.Field,
	)
}
