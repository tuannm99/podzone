package activity

type ActivityProjection interface {
	ActivityCommandProjection
	ActivityQueryRepository
}
