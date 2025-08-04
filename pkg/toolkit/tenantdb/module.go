package tenantdb

// Package tenantdb provides functionality for managing multi-tenant database connections
// with support for different tenant isolation strategies.
//
// Key design:
// There are two types of tenant database isolation:
//   1. Group-based (separate schema per tenant) - used for cold/warm tenants
//      Database name: group_XX
//      Schema structure:
//        - tenant1_ + table
//        - tenant2_ + table
//        - tenantN_ + table
//        ...
//
//   2. Per-tenant database - used for hot tenants
//      Database name: tenant_XX
//      Schema structure:
//        - default_ + table
//        ...
//
// The package automatically handles:
// - Database connection pooling
// - Schema creation and management
// - Tenant context propagation
// - Connection lifecycle management

// with seperate schema
// if we created one connection for many tenant if a thousand, it may not ok
// -> instead we create pool of conenction for each tenant, but max pool size is 200, with local LRU cached
// when another tenant not in the pool, remove the latest and switch to that tenant, keep connection,
// change the schema (switching)

// with seperate database
// -> max 50 connection per application -> inside one namespace of kubernetes
// assume that we have 1000 hot tenant -> we have 1000/50 => 20 namespace
// the namespace provided at env level of each pod inside kubernetes
