package inputport

// IAMUsecase composes the bounded IAM usecase slices.
type IAMUsecase interface {
	TenantUsecase
	OrganizationUsecase
	PolicyUsecase
	GroupUsecase
	PlatformPrincipalUsecase
	AuthzUsecase
}
