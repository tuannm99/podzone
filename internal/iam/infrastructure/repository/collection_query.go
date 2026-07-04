package repository

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/pkg/collection"
)

type collectionColumn struct {
	sql  string
	text bool
}

var organizationCollectionColumns = map[string]collectionColumn{
	"id":        {sql: "organization.id", text: true},
	"slug":      {sql: "organization.slug", text: true},
	"name":      {sql: "organization.name", text: true},
	"createdAt": {sql: "organization.created_at"},
	"updatedAt": {sql: "organization.updated_at"},
}

var policyCollectionColumns = map[string]collectionColumn{
	"id":             {sql: "id"},
	"scope":          {sql: "scope", text: true},
	"organizationId": {sql: "org_id", text: true},
	"name":           {sql: "name", text: true},
	"description":    {sql: "description", text: true},
	"isSystem":       {sql: "is_system"},
	"defaultVersion": {sql: "default_version", text: true},
	"createdAt":      {sql: "created_at"},
	"updatedAt":      {sql: "updated_at"},
}

var groupCollectionColumns = map[string]collectionColumn{
	"id":             {sql: "id"},
	"scope":          {sql: "scope", text: true},
	"organizationId": {sql: "org_id", text: true},
	"tenantId":       {sql: "tenant_id", text: true},
	"name":           {sql: "name", text: true},
	"description":    {sql: "description", text: true},
	"isSystem":       {sql: "is_system"},
	"createdAt":      {sql: "created_at"},
	"updatedAt":      {sql: "updated_at"},
}

var permissionCollectionColumns = map[string]collectionColumn{
	"id":       {sql: "id"},
	"name":     {sql: "name", text: true},
	"resource": {sql: "resource", text: true},
	"action":   {sql: "action", text: true},
}

var policyVersionCollectionColumns = map[string]collectionColumn{
	"id":        {sql: "pv.id"},
	"version":   {sql: "pv.version", text: true},
	"isDefault": {sql: "pv.is_default"},
	"createdAt": {sql: "pv.created_at"},
}

var managedPolicyCollectionColumns = map[string]collectionColumn{
	"id":             {sql: "p.id"},
	"scope":          {sql: "p.scope", text: true},
	"organizationId": {sql: "p.org_id", text: true},
	"name":           {sql: "p.name", text: true},
	"description":    {sql: "p.description", text: true},
	"isSystem":       {sql: "p.is_system"},
	"defaultVersion": {sql: "p.default_version", text: true},
	"createdAt":      {sql: "p.created_at"},
	"updatedAt":      {sql: "p.updated_at"},
}

var groupMemberCollectionColumns = map[string]collectionColumn{
	"userId":    {sql: "gm.user_id"},
	"createdAt": {sql: "gm.created_at"},
}

var inlinePolicyCollectionColumns = map[string]collectionColumn{
	"name":        {sql: "ip.name", text: true},
	"description": {sql: "ip.description", text: true},
	"createdAt":   {sql: "ip.created_at"},
	"updatedAt":   {sql: "ip.updated_at"},
}

var policyAttachmentCollectionColumns = map[string]collectionColumn{
	"attachmentType": {sql: "attachment_type", text: true},
	"scope":          {sql: "scope", text: true},
	"tenantId":       {sql: "tenant_id", text: true},
	"roleId":         {sql: "role_id"},
	"roleName":       {sql: "role_name", text: true},
	"userId":         {sql: "user_id"},
	"groupId":        {sql: "group_id"},
	"groupName":      {sql: "group_name", text: true},
	"createdAt":      {sql: "created_at"},
}

var tenantMembershipCollectionColumns = map[string]collectionColumn{
	"tenantId":  {sql: "tm.tenant_id", text: true},
	"userId":    {sql: "tm.user_id"},
	"roleId":    {sql: "tm.role_id"},
	"roleName":  {sql: "r.name", text: true},
	"status":    {sql: "tm.status", text: true},
	"createdAt": {sql: "tm.created_at"},
	"updatedAt": {sql: "tm.updated_at"},
}

var organizationMembershipCollectionColumns = map[string]collectionColumn{
	"organizationId": {sql: "om.org_id", text: true},
	"userId":         {sql: "om.user_id"},
	"roleId":         {sql: "om.role_id"},
	"roleName":       {sql: "r.name", text: true},
	"status":         {sql: "om.status", text: true},
	"createdAt":      {sql: "om.created_at"},
	"updatedAt":      {sql: "om.updated_at"},
}

var tenantInviteCollectionColumns = map[string]collectionColumn{
	"id":              {sql: "ti.id", text: true},
	"tenantId":        {sql: "ti.tenant_id", text: true},
	"email":           {sql: "ti.email", text: true},
	"roleId":          {sql: "ti.role_id"},
	"roleName":        {sql: "r.name", text: true},
	"status":          {sql: "ti.status", text: true},
	"invitedByUserId": {sql: "ti.invited_by_user_id"},
	"expiresAt":       {sql: "ti.expires_at"},
	"createdAt":       {sql: "ti.created_at"},
	"updatedAt":       {sql: "ti.updated_at"},
}

var platformRoleCollectionColumns = map[string]collectionColumn{
	"userId":    {sql: "upr.user_id"},
	"roleId":    {sql: "upr.role_id"},
	"roleName":  {sql: "r.name", text: true},
	"status":    {sql: "upr.status", text: true},
	"createdAt": {sql: "upr.created_at"},
	"updatedAt": {sql: "upr.updated_at"},
}

func listIAMCollectionModels[M any](
	ctx context.Context,
	db *sqlx.DB,
	query collection.Query,
	from string,
	selectColumns []string,
	fixedWhere []sq.Sqlizer,
	columns map[string]collectionColumn,
	searchColumns []string,
	defaultSort string,
	stableSort string,
) (collection.Page[M], error) {
	normalized, dynamicWhere, orderBy, err := buildIAMCollectionQuery(
		query,
		columns,
		searchColumns,
		defaultSort,
	)
	if err != nil {
		return collection.Page[M]{}, err
	}
	where := append(append([]sq.Sqlizer{}, fixedWhere...), dynamicWhere...)
	countBuilder := sq.Select("COUNT(*)").From(from)
	for _, predicate := range where {
		countBuilder = countBuilder.Where(predicate)
	}
	countSQL, countArgs, err := countBuilder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return collection.Page[M]{}, err
	}
	var total int64
	if err := db.GetContext(ctx, &total, countSQL, countArgs...); err != nil {
		return collection.Page[M]{}, err
	}

	listBuilder := sq.Select(selectColumns...).From(from)
	for _, predicate := range where {
		listBuilder = listBuilder.Where(predicate)
	}
	listSQL, listArgs, err := listBuilder.
		OrderBy(orderBy, stableSort).
		Limit(uint64(normalized.PageSize)).
		Offset(uint64(normalized.Offset())).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return collection.Page[M]{}, err
	}
	var rows []M
	if err := db.SelectContext(ctx, &rows, listSQL, listArgs...); err != nil {
		return collection.Page[M]{}, err
	}
	return collection.NewPage(rows, total, normalized), nil
}

func buildIAMCollectionQuery(
	query collection.Query,
	columns map[string]collectionColumn,
	searchColumns []string,
	defaultSort string,
) (collection.Query, []sq.Sqlizer, string, error) {
	normalized := query.Normalize()
	where := make([]sq.Sqlizer, 0, len(normalized.Filters)+1)
	if search := strings.TrimSpace(normalized.Search); search != "" {
		pattern := "%" + escapeIAMLike(search) + "%"
		searchClauses := make(sq.Or, 0, len(searchColumns))
		for _, column := range searchColumns {
			searchClauses = append(searchClauses, likeIAMColumn(column, pattern))
		}
		where = append(where, searchClauses)
	}
	for _, filter := range normalized.Filters {
		column, ok := columns[filter.Field]
		if !ok {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported IAM filter field %q",
				collection.ErrInvalidQuery,
				filter.Field,
			)
		}
		clause, err := iamFilterClause(column, filter)
		if err != nil {
			return collection.Query{}, nil, "", err
		}
		where = append(where, clause)
	}
	sortColumn := defaultSort
	if normalized.SortBy != "" {
		column, ok := columns[normalized.SortBy]
		if !ok {
			return collection.Query{}, nil, "", fmt.Errorf(
				"%w: unsupported IAM sort field %q",
				collection.ErrInvalidQuery,
				normalized.SortBy,
			)
		}
		sortColumn = column.sql
	}
	direction := "DESC"
	if normalized.SortDirection == collection.SortAscending {
		direction = "ASC"
	}
	return normalized, where, sortColumn + " " + direction, nil
}

func iamFilterClause(column collectionColumn, filter collection.Filter) (sq.Sqlizer, error) {
	if len(filter.Values) == 0 {
		return nil, fmt.Errorf(
			"%w: IAM filter %q requires a value",
			collection.ErrInvalidQuery,
			filter.Field,
		)
	}
	switch filter.Operator {
	case collection.FilterEqual:
		return sq.Eq{column.sql: filter.Values[0]}, nil
	case collection.FilterNotEqual:
		return sq.NotEq{column.sql: filter.Values[0]}, nil
	case collection.FilterContains:
		if !column.text {
			return nil, unsupportedIAMFilterOperator(filter)
		}
		return likeIAMColumn(column.sql, "%"+escapeIAMLike(filter.Values[0])+"%"), nil
	case collection.FilterStartsWith:
		if !column.text {
			return nil, unsupportedIAMFilterOperator(filter)
		}
		return likeIAMColumn(column.sql, escapeIAMLike(filter.Values[0])+"%"), nil
	case collection.FilterGreaterThan:
		return sq.Gt{column.sql: filter.Values[0]}, nil
	case collection.FilterGreaterThanOrEqual:
		return sq.GtOrEq{column.sql: filter.Values[0]}, nil
	case collection.FilterLessThan:
		return sq.Lt{column.sql: filter.Values[0]}, nil
	case collection.FilterLessThanOrEqual:
		return sq.LtOrEq{column.sql: filter.Values[0]}, nil
	case collection.FilterIn:
		return sq.Eq{column.sql: filter.Values}, nil
	default:
		return nil, unsupportedIAMFilterOperator(filter)
	}
}

func unsupportedIAMFilterOperator(filter collection.Filter) error {
	return fmt.Errorf(
		"%w: unsupported IAM filter operator %q for field %q",
		collection.ErrInvalidQuery,
		filter.Operator,
		filter.Field,
	)
}

func likeIAMColumn(column string, pattern string) sq.Sqlizer {
	return sq.Expr("LOWER("+column+") LIKE LOWER(?) ESCAPE '\\'", pattern)
}

func escapeIAMLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
