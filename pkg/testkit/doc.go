// Package testkit contains integration-test helpers shared across services.
//
// Helpers start process-wide testcontainers for common dependencies and return
// ready-to-use connection information. If Docker is unavailable, tests are
// skipped instead of failing so unit test suites can still run in restricted
// environments.
//
// Use testkit for repository and infrastructure integration tests. Domain and
// interactor tests should prefer mockery-generated mocks or small in-memory
// fakes only when the fake is a reusable testing component.
package testkit
