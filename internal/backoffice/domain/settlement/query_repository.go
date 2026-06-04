package settlement

import (
	"context"
)

type SettlementRecordQueryRepository interface {
	GetSettlementRecord(ctx context.Context, storeID string, orderID string) (*SettlementRecord, error)
}
