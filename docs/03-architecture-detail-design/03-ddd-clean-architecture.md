# DDD And Clean Architecture Conventions

## Purpose

This document defines how Podzone applies Domain-Driven Design while keeping Clean Architecture boundaries.

It is the reference pattern for refactoring Backoffice and IAM into feature-owned domains without turning handlers, repositories, or shared packages into business-rule containers.

## Core Rule

DDD names the business model.

Clean Architecture controls dependency direction.

That means:

- domain code owns business language, state transitions, invariants, and domain errors
- input ports expose usecase contracts
- output ports expose dependencies needed by the domain/application layer
- interactors orchestrate usecases and transactions
- controllers only map transport requests to commands/queries
- infrastructure implements output ports
- cross-service communication uses adapters or events, not direct domain imports

## Package Shape

Use this shape for a simple service context:

```text
internal/<service>/domain/<context>/
  entity/
  inputport/
  outputport/
  interactor.go
```

Use this shape when a context grows:

```text
internal/<service>/domain/<context>/
  entity/
    <aggregate>.go
    events.go
    errors.go
  inputport/
    command.go
    query.go
  outputport/
    repository.go
    publisher.go
  interactor/
    command.go
    query.go
    policy.go
```

Transport and infrastructure stay outside the domain:

```text
internal/<service>/controller/<transport>/<feature>/
internal/<service>/controller/mapper/
internal/<service>/controller/eventhandler/<consumer>/
internal/<service>/infrastructure/repository/<context>/
internal/<service>/infrastructure/messaging/<consumer-or-publisher>/
internal/<service>/infrastructure/<external-client>/
```

## Backoffice Target Shape

Backoffice should evolve toward:

```text
internal/backoffice/domain/
  workspace/
  catalog/
  order/
  routing/
  fulfillment/
  exception/
  settlement/
  activity/
```

The current GraphQL contract can stay stable while internals split.

The existing `RoutedOrder` GraphQL type should become a composed read model, not a target aggregate.

## Entity And Aggregate Pattern

Entities are plain domain types.

Aggregates should expose behavior methods for state transitions instead of letting handlers or repositories mutate business state.

Use private fields when an invariant must be protected by aggregate methods.

Preferred:

```go
func (o *FulfillmentOrder) MarkInTransit(carrier string, trackingNumber string, now time.Time) error {
    if o.PartnerID == "" {
        return ErrPartnerNotAssigned
    }
    if trackingNumber == "" {
        return ErrTrackingNumberRequired
    }
    o.ShipmentStatus = ShipmentStatusInTransit
    o.Carrier = carrier
    o.TrackingNumber = trackingNumber
    o.UpdatedAt = now
    o.record(FulfillmentOrderInTransit{OrderID: o.ID, Carrier: carrier})
    return nil
}
```

Avoid:

```go
order.ShipmentStatus = "in_transit"
order.TrackingNumber = req.TrackingNumber
```

inside a GraphQL resolver or repository.

## Value Object Pattern

Value objects have no identity and are compared by value.

Use value objects for concepts like permission action/resource pairs, money, region codes, tracking numbers, and policy conditions when validation matters.

Preferred:

```go
type Money struct {
    currency string
    cents int64
}

func NewMoney(currency string, cents int64) (Money, error) {
    if currency == "" {
        return Money{}, ErrInvalidCurrency
    }
    if cents < 0 {
        return Money{}, ErrInvalidMoneyAmount
    }
    return Money{currency: currency, cents: cents}, nil
}
```

Avoid passing unvalidated strings through multiple layers when the business concept has rules.

## Rehydration Pattern

Repositories load aggregate roots from storage through snapshot/rehydration functions.

Rehydration must validate enough to prevent invalid domain objects, but it must not emit domain events.

Example:

```go
type FulfillmentOrderSnapshot struct {
    ID string
    StoreID string
    PartnerID string
    ShipmentStatus string
}

func RehydrateFulfillmentOrder(snapshot FulfillmentOrderSnapshot) (*FulfillmentOrder, error) {
    if snapshot.ID == "" {
        return nil, ErrInvalidFulfillmentOrderID
    }
    return &FulfillmentOrder{
        id: snapshot.ID,
        storeID: snapshot.StoreID,
        partnerID: snapshot.PartnerID,
        shipmentStatus: snapshot.ShipmentStatus,
    }, nil
}
```

Factories like `NewFulfillmentOrder` may emit events.

Rehydration functions must not.

## Domain Events

Domain events represent facts that already happened.

Use past-tense names:

- `CustomerOrderReceived`
- `PartnerRecommended`
- `FulfillmentPartnerAssigned`
- `ShipmentDelivered`
- `SettlementReconciled`

Event structs belong near the aggregate that emits them:

```text
domain/fulfillment/entity/events.go
```

An aggregate may keep pending events:

```go
type FulfillmentOrder struct {
    ID string
    pendingEvents []DomainEvent
}

func (o *FulfillmentOrder) PullEvents() []DomainEvent {
    events := o.pendingEvents
    o.pendingEvents = nil
    return events
}
```

Use outbox only when events must cross service boundaries or survive process failure.

For in-process coordination during migration, return events from the interactor or publish through an output port with a transaction boundary.

## Commands And Queries

Commands express intent to change state.

Queries express read intent and must not mutate state.

Use explicit names:

```go
type AssignFulfillmentPartnerCmd struct {
    StoreID string
    OrderID string
    PartnerID string
    ActorUserID uint
}

type ListStoreOrdersQuery struct {
    StoreID string
    Status string
}
```

Input ports should split command and query surfaces when the context has meaningful read/write differences:

```go
type FulfillmentCommandUsecase interface {
    AssignPartner(ctx context.Context, cmd AssignFulfillmentPartnerCmd) (*entity.FulfillmentOrder, error)
    UpdateShipment(ctx context.Context, cmd UpdateShipmentCmd) (*entity.FulfillmentOrder, error)
}

type FulfillmentQueryUsecase interface {
    GetFulfillmentOrder(ctx context.Context, query GetFulfillmentOrderQuery) (*entity.FulfillmentOrderView, error)
    ListFulfillmentQueue(ctx context.Context, query ListFulfillmentQueueQuery) ([]entity.FulfillmentQueueItem, error)
}
```

## Interactor Pattern

Interactors orchestrate:

- authorization assumptions passed from the controller
- loading aggregates through output ports
- invoking aggregate behavior
- saving through output ports
- appending/publishing domain events
- building read models when the query is context-local

Interactors should not:

- parse GraphQL/gRPC requests
- know HTTP status codes
- return protobuf or GraphQL generated types
- import repositories directly from another context
- contain SQL

## Transaction Boundary

Usecases/interactors own transaction boundaries.

Domain entities do not know transactions.

Infrastructure implements transaction execution behind an output port:

```go
type TxManager interface {
    WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
```

Command usecases that save aggregates and outbox events should do both inside the same transaction when consistency matters.

The transaction manager belongs in an output port or application port. Do not pass `*sqlx.DB`, `*gorm.DB`, or Mongo clients into domain/interactor code unless the package is explicitly infrastructure.

## Output Port Pattern

Repositories are defined in the context that consumes them:

```go
type FulfillmentOrderRepository interface {
    Get(ctx context.Context, storeID string, orderID string) (*entity.FulfillmentOrder, error)
    Save(ctx context.Context, order *entity.FulfillmentOrder) error
}
```

Use separate command/query repositories when it helps dependency control:

```go
type FulfillmentCommandRepository interface {
    Save(ctx context.Context, order *entity.FulfillmentOrder) error
}

type FulfillmentQueryRepository interface {
    Get(ctx context.Context, storeID string, orderID string) (*entity.FulfillmentOrder, error)
    ListQueue(ctx context.Context, storeID string) ([]entity.FulfillmentQueueItem, error)
}
```

Infrastructure repositories may share a concrete SQL implementation, but the module wiring should expose only the interface required by the current runtime.

Do not expose generic CRUD repositories at domain/application boundaries.

Avoid:

```go
type Repository[T any] interface {
    Create(ctx context.Context, item T) error
    Update(ctx context.Context, item T) error
    FindByID(ctx context.Context, id string) (*T, error)
}
```

Prefer aggregate-specific repositories whose methods reflect the business access pattern.

Generic helpers are acceptable inside infrastructure only.

## Controller Pattern

Controllers are inbound adapters.

They should:

- extract auth/session/store scope
- validate transport-level required fields
- map request DTOs to command/query structs
- call one usecase
- map domain result to response DTO
- translate domain errors to transport errors

They should not:

- mutate aggregate fields directly
- call repositories
- call another service domain/interactor directly
- implement routing, settlement, or fulfillment business rules

## Mapper Pattern

Transport mapping belongs in `controller/mapper`.

Mapper functions may convert:

- GraphQL generated input to command/query
- domain entity/view to GraphQL generated model
- protobuf request/response to command/query/entity

Mapper functions must not:

- call repositories
- call usecases
- make authorization decisions
- publish events

## Domain Service And Policy Pattern

Use a domain service when a rule spans multiple aggregates but remains inside one bounded context.

Use a process policy when an event in one context triggers work in another context.

Examples:

- `RoutingPolicy` decides whether an order can be auto-assigned.
- `SettlementPolicy` recalculates margin from cost inputs.
- `FulfillmentIssuePolicy` opens an exception when shipment status reports a delivery issue.

Policy naming should be business-facing, not technical.

## Cross-Context Rules

Inside the same service, one context must not import another context's repository or interactor.

Allowed communication:

- input ports when one context needs a synchronous application service
- domain events / process policies for workflow propagation
- read models/projections when the source context explicitly owns the projection

Forbidden:

```go
import "internal/backoffice/infrastructure/repository/routing"
```

from `domain/settlement`.

Forbidden:

```go
import "internal/backoffice/domain/routing/interactor"
```

from `domain/fulfillment`.

## Module Wiring

Fx modules should mirror context boundaries:

```text
internal/backoffice/domain/fulfillment/module.go
internal/backoffice/domain/settlement/module.go
internal/backoffice/infrastructure/repository/fulfillment/module.go
internal/backoffice/infrastructure/repository/settlement/module.go
```

When command/query split is real, expose modules explicitly:

```go
var CommandModule = fx.Options(...)
var QueryModule = fx.Options(...)
var Module = fx.Options(CommandModule, QueryModule)
```

The all-in-one module may stay for current deployability, but the split modules should compile independently.

Name modules by their real dependency graph.

If a command handler needs query dependencies for authorization or read-after-write behavior, do not call the runtime dependency graph command-only. Use names such as:

- `CommandRuntimeModule`
- `CommandWithQueryModule`
- `CQRSRuntimeModule`

Every struct provided through `fx.As(...)` must have a compile-time assertion near the struct:

```go
var _ outputport.FulfillmentCommandRepository = (*FulfillmentRepository)(nil)
```

Constructors should return concrete types by default:

```go
func NewInteractor(p InteractorParams) *Interactor
```

Use `fx.As(...)` at module boundaries to expose interfaces.

Avoid returning interfaces from constructors unless the constructor intentionally hides multiple concrete implementations.

Add `fx.ValidateApp` tests for deployable modules and important split modules so missing dependencies fail in tests, not in Docker startup.

## DTO Boundaries

Keep DTOs separated by layer:

- transport DTO: protobuf, GraphQL generated model, HTTP request/response
- application DTO: command/query/result/view structs
- domain model: aggregate, entity, value object, domain event
- persistence DTO: SQL row, Mongo document, projection row

Do not pass protobuf, GraphQL generated models, SQL rows, or Mongo documents into domain methods.

Query usecases may return read models/views without reconstructing full aggregates when no invariant is being enforced.

Command usecases should use aggregates when state transitions or invariants matter.

## Import Rules

Domain packages must not import:

- `controller`
- `infrastructure`
- gRPC/HTTP frameworks
- SQL/Mongo/Kafka/Redis clients
- generated protobuf or GraphQL transport types

Interactor/application packages must not import:

- `controller`
- concrete infrastructure packages from another context
- transport status/error packages such as `google.golang.org/grpc/status`

Infrastructure may import domain and output ports.

Controllers may import input ports, mappers, generated transport types, and transport error packages.

## Testing Pattern

Use table-driven domain tests for aggregate invariants and interactor workflows, If the business too complicated or have too much testcase, split to many test.

Use testify/mockery-generated mocks for output ports.

Do not write hand fakes for new tests unless the fake is a reusable testkit implementation.

Minimum useful tests:

- aggregate state transition test
- interactor command success test
- interactor validation/permission failure test
- controller mapping/error test only when transport behavior is non-trivial
- Fx graph validation test for deployable modules

## Migration Checklist

Before moving code into a DDD context:

1. Name the business capability.
2. Identify commands and events.
3. Identify the aggregate that owns the invariant.
4. Define input ports.
5. Define output ports.
6. Move business rules into aggregate/interactor.
7. Keep transport mapping in controller/mapper.
8. Wire infrastructure through Fx modules.
9. Add tests around aggregate/interactor behavior.
10. Add compile-time assertions for every new `fx.As(...)` provider.
11. Add or update `fx.ValidateApp` tests for changed runtime modules.
12. Run `gofumpt`, `golangci-lint`, and focused tests before handing off.
