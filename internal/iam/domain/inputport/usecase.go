package inputport

// IAMUsecase composes the bounded IAM usecase slices.
type IAMUsecase interface {
	IAMCommandUsecase
	IAMQueryUsecase
}

// IAMLegacyUsecase keeps the pre-CQRS bounded slices visible during migration.
type IAMLegacyUsecase interface {
	TenantUsecase
	OrganizationUsecase
	PolicyUsecase
	GroupUsecase
	PlatformPrincipalUsecase
	AuthzUsecase
}
