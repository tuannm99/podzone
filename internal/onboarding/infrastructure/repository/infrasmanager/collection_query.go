package repository

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tuannm99/podzone/pkg/collection"
)

type collectionFieldKind uint8

const (
	collectionFieldString collectionFieldKind = iota
	collectionFieldInt
	collectionFieldBool
	collectionFieldTime
)

type collectionField struct {
	path string
	kind collectionFieldKind
}

var connectionCollectionFields = map[string]collectionField{
	"tenantId":  {path: "tenant_id", kind: collectionFieldString},
	"infraType": {path: "infra_type", kind: collectionFieldString},
	"name":      {path: "name", kind: collectionFieldString},
	"endpoint":  {path: "endpoint", kind: collectionFieldString},
	"secretRef": {path: "secret_ref", kind: collectionFieldString},
	"status":    {path: "status", kind: collectionFieldString},
	"version":   {path: "version", kind: collectionFieldInt},
	"createdAt": {path: "created_at", kind: collectionFieldTime},
	"updatedAt": {path: "updated_at", kind: collectionFieldTime},
	"deletedAt": {path: "deleted_at", kind: collectionFieldTime},
}

var eventCollectionFields = map[string]collectionField{
	"id":            {path: "id", kind: collectionFieldString},
	"correlationId": {path: "correlation_id", kind: collectionFieldString},
	"tenantId":      {path: "tenant_id", kind: collectionFieldString},
	"infraType":     {path: "infra_type", kind: collectionFieldString},
	"name":          {path: "name", kind: collectionFieldString},
	"action":        {path: "action", kind: collectionFieldString},
	"status":        {path: "status", kind: collectionFieldString},
	"error":         {path: "error", kind: collectionFieldString},
	"createdAt":     {path: "created_at", kind: collectionFieldTime},
}

var databaseClusterCollectionFields = map[string]collectionField{
	"name":               {path: "name", kind: collectionFieldString},
	"engine":             {path: "engine", kind: collectionFieldString},
	"region":             {path: "region", kind: collectionFieldString},
	"placementDb":        {path: "placement_db", kind: collectionFieldString},
	"maxTenants":         {path: "max_tenants", kind: collectionFieldInt},
	"currentTenants":     {path: "current_tenants", kind: collectionFieldInt},
	"maxSchemas":         {path: "max_schemas", kind: collectionFieldInt},
	"currentSchemas":     {path: "current_schemas", kind: collectionFieldInt},
	"maxConnections":     {path: "max_connections", kind: collectionFieldInt},
	"currentConnections": {path: "current_connections", kind: collectionFieldInt},
	"status":             {path: "status", kind: collectionFieldString},
	"healthy":            {path: "healthy", kind: collectionFieldBool},
	"createdAt":          {path: "created_at", kind: collectionFieldTime},
	"updatedAt":          {path: "updated_at", kind: collectionFieldTime},
}

var kubernetesClusterCollectionFields = map[string]collectionField{
	"name":      {path: "name", kind: collectionFieldString},
	"region":    {path: "region", kind: collectionFieldString},
	"status":    {path: "status", kind: collectionFieldString},
	"healthy":   {path: "healthy", kind: collectionFieldBool},
	"createdAt": {path: "created_at", kind: collectionFieldTime},
	"updatedAt": {path: "updated_at", kind: collectionFieldTime},
}

var runtimePoolCollectionFields = map[string]collectionField{
	"name":           {path: "name", kind: collectionFieldString},
	"kind":           {path: "kind", kind: collectionFieldString},
	"maxTenants":     {path: "max_tenants", kind: collectionFieldInt},
	"currentTenants": {path: "current_tenants", kind: collectionFieldInt},
	"status":         {path: "status", kind: collectionFieldString},
	"healthy":        {path: "healthy", kind: collectionFieldBool},
	"createdAt":      {path: "created_at", kind: collectionFieldTime},
	"updatedAt":      {path: "updated_at", kind: collectionFieldTime},
}

func buildConnectionCollection(
	tenantID string,
	includeDeleted bool,
	query collection.Query,
) (collection.Query, bson.M, bson.D, error) {
	clauses := bson.A{bson.M{"tenant_id": tenantID}}
	if !includeDeleted {
		clauses = append(clauses, bson.M{"deleted_at": nil})
	}
	return buildInfrastructureCollection(
		query,
		clauses,
		connectionCollectionFields,
		[]string{"infraType", "name", "endpoint", "secretRef", "status"},
		"updatedAt",
		"name",
	)
}

func buildEventCollection(
	tenantID string,
	query collection.Query,
) (collection.Query, bson.M, bson.D, error) {
	return buildInfrastructureCollection(
		query,
		bson.A{bson.M{"tenant_id": tenantID}},
		eventCollectionFields,
		[]string{"id", "correlationId", "infraType", "name", "action", "status", "error"},
		"createdAt",
		"id",
	)
}

func buildDatabaseClusterCollection(
	query collection.Query,
) (collection.Query, bson.M, bson.D, error) {
	return buildInfrastructureCollection(
		query,
		bson.A{bson.M{}},
		databaseClusterCollectionFields,
		[]string{"name", "engine", "region", "placementDb", "status"},
		"updatedAt",
		"name",
	)
}

func buildKubernetesClusterCollection(
	query collection.Query,
) (collection.Query, bson.M, bson.D, error) {
	return buildInfrastructureCollection(
		query,
		bson.A{bson.M{}},
		kubernetesClusterCollectionFields,
		[]string{"name", "region", "status"},
		"updatedAt",
		"name",
	)
}

func buildRuntimePoolCollection(
	query collection.Query,
) (collection.Query, bson.M, bson.D, error) {
	return buildInfrastructureCollection(
		query,
		bson.A{bson.M{}},
		runtimePoolCollectionFields,
		[]string{"name", "kind", "status"},
		"updatedAt",
		"name",
	)
}

func buildInfrastructureCollection(
	query collection.Query,
	clauses bson.A,
	fields map[string]collectionField,
	searchFields []string,
	defaultSort string,
	stableSort string,
) (collection.Query, bson.M, bson.D, error) {
	normalized := query.Normalize()
	if search := strings.TrimSpace(normalized.Search); search != "" {
		pattern := primitive.Regex{Pattern: regexp.QuoteMeta(search), Options: "i"}
		searchClauses := make(bson.A, 0, len(searchFields))
		for _, fieldName := range searchFields {
			searchClauses = append(searchClauses, bson.M{fields[fieldName].path: pattern})
		}
		clauses = append(clauses, bson.M{"$or": searchClauses})
	}
	for _, filter := range normalized.Filters {
		clause, err := infrastructureFilter(fields, filter)
		if err != nil {
			return collection.Query{}, nil, nil, err
		}
		clauses = append(clauses, clause)
	}

	sortFieldName := defaultSort
	if normalized.SortBy != "" {
		sortFieldName = normalized.SortBy
	}
	sortField, ok := fields[sortFieldName]
	if !ok {
		return collection.Query{}, nil, nil, fmt.Errorf(
			"%w: unsupported infrastructure sort field %q",
			collection.ErrInvalidQuery,
			sortFieldName,
		)
	}
	direction := -1
	if normalized.SortDirection == collection.SortAscending {
		direction = 1
	}
	sort := bson.D{{Key: sortField.path, Value: direction}}
	if stableField := fields[stableSort]; stableField.path != sortField.path {
		sort = append(sort, bson.E{Key: stableField.path, Value: 1})
	}
	return normalized, bson.M{"$and": clauses}, sort, nil
}

func infrastructureFilter(
	fields map[string]collectionField,
	filter collection.Filter,
) (bson.M, error) {
	field, ok := fields[filter.Field]
	if !ok {
		return nil, fmt.Errorf(
			"%w: unsupported infrastructure filter field %q",
			collection.ErrInvalidQuery,
			filter.Field,
		)
	}
	if len(filter.Values) == 0 {
		return nil, fmt.Errorf(
			"%w: filter %q requires a value",
			collection.ErrInvalidQuery,
			filter.Field,
		)
	}
	if filter.Operator == collection.FilterContains || filter.Operator == collection.FilterStartsWith {
		if field.kind != collectionFieldString {
			return nil, unsupportedInfrastructureFilter(filter)
		}
		pattern := regexp.QuoteMeta(filter.Values[0])
		if filter.Operator == collection.FilterStartsWith {
			pattern = "^" + pattern
		}
		return bson.M{field.path: primitive.Regex{Pattern: pattern, Options: "i"}}, nil
	}
	values, err := convertInfrastructureValues(field, filter.Values)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: invalid value for field %q",
			collection.ErrInvalidQuery,
			filter.Field,
		)
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
		return nil, unsupportedInfrastructureFilter(filter)
	}
}

func convertInfrastructureValues(field collectionField, values []string) (bson.A, error) {
	out := make(bson.A, 0, len(values))
	for _, value := range values {
		switch field.kind {
		case collectionFieldInt:
			number, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}
			out = append(out, number)
		case collectionFieldBool:
			boolean, err := strconv.ParseBool(value)
			if err != nil {
				return nil, err
			}
			out = append(out, boolean)
		case collectionFieldTime:
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

func unsupportedInfrastructureFilter(filter collection.Filter) error {
	return fmt.Errorf(
		"%w: unsupported operator %q for infrastructure field %q",
		collection.ErrInvalidQuery,
		filter.Operator,
		filter.Field,
	)
}
