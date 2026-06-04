package settlement

type SettlementRecordRepository interface {
	SettlementRecordCommandRepository
	SettlementRecordQueryRepository
}
