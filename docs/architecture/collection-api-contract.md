# Collection API Contract

## Purpose

Every unbounded list in Podzone uses server-side pagination, search, filtering,
and sorting. Frontends must not fetch a full collection and paginate it only in
the browser.

## Canonical Semantics

Collection requests contain:

- `page`: one-based page number; defaults to `1`
- `page_size`: defaults to `20`, capped at `100`
- `search`: free text applied only to resource-owned searchable fields
- `filters`: typed field, operator, and values
- `sort_by`: a resource-owned sortable field
- `sort_direction`: `ASC` or `DESC`

Collection responses contain:

- the resource-specific `items` field
- `page_info.total`
- `page_info.page`
- `page_info.page_size`
- `page_info.total_pages`
- `page_info.has_next`
- `page_info.has_previous`

The protobuf wire types live in `common/v1/common.proto`. The GraphQL
equivalents live in `controller/graphql/schema/common.graphqls`. Application
and repository boundaries use `pkg/collection` and do not depend on transport
types.

## Search And Filter Safety

`search` does not mean arbitrary SQL or arbitrary document traversal. Each
resource declares searchable fields.

Repositories must:

- map public field names to known storage columns
- whitelist supported operators per field
- parameterize all values
- reject unsupported field, operator, and sort combinations
- avoid returning secret or internal-only fields through search

Client-provided storage column names and raw query expressions are forbidden.

## Frontend Pattern

Collection pages keep `page`, `pageSize`, `search`, filters, and sort in typed
route search state when the state is shareable.

Remote reads use a resource/query primitive keyed by collection state. Changing
page, search, filter, or sort triggers a backend request. The UI renders only
the returned page and uses `page_info` for controls.

Client-only pagination is allowed only for explicitly bounded static data.

## Migration Status

Implemented:

- Auth sessions
- Auth audit logs
- shared protobuf, application, GraphQL, and frontend contracts

Pending migration:

- IAM organizations, policies, versions, attachments, groups, principals,
  memberships, roles, and invites
- onboarding workspaces, stores, provisioning requests, connections, events,
  placements, and resource inventories
- backoffice stores, catalog drafts/candidates, orders, partners, activity,
  fulfillment, and settlement views

An endpoint remains pending until its repository performs pagination and
filtering before materializing rows. Adding browser pagination alone does not
complete the migration.
