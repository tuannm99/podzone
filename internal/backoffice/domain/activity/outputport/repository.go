package outputport

type ActivityProjection interface {
	ActivityCommandProjection
	ActivityQueryRepository
}
