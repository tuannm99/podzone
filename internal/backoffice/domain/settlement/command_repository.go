package settlement

import (
	"context"
)

type SettlementRecordCommandRepository interface {
	SaveSettlementRecord(ctx context.Context, record *SettlementRecord) error
}
