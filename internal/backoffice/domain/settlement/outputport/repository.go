package outputport

type SettlementRecordRepository interface {
	SettlementRecordCommandRepository
	SettlementRecordQueryRepository
}
